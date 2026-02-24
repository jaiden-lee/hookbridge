package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hookbridge/gen/tunnelv1"
	"hookbridge/internal/api"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ConnectConfig struct {
	Name      string
	LocalPort int
	ServerURL string
}

func RunConnect(ctx context.Context, cfg ConnectConfig) error {
	if strings.TrimSpace(cfg.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if cfg.LocalPort <= 0 || cfg.LocalPort > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if strings.TrimSpace(cfg.ServerURL) == "" {
		return fmt.Errorf("server URL cannot be empty")
	}

	connectResp, err := connectOrCreateTunnel(ctx, cfg)
	if err != nil {
		return err
	}

	grpcAddress := fmt.Sprintf("%s:%d", connectResp.TunnelIp, connectResp.Port)
	conn, err := grpc.NewClient(
		grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to dial tunnel grpc endpoint %s: %w", grpcAddress, err)
	}
	defer conn.Close()

	grpcClient := tunnelv1.NewTunnelServiceClient(conn)
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	md := metadata.Pairs("tunnel-name", strings.ToLower(cfg.Name))
	streamCtx = metadata.NewOutgoingContext(streamCtx, md)

	stream, err := grpcClient.OpenTunnel(streamCtx)
	if err != nil {
		return fmt.Errorf("failed to open tunnel stream: %w", err)
	}
	defer stream.CloseSend()

	log.Printf("connected to tunnel %s at %s", cfg.Name, grpcAddress)

	requestCh := make(chan *tunnelv1.HttpRequest, 64)
	responseCh := make(chan *tunnelv1.HttpResponse, 64)
	workerDone := make(chan struct{})
	doneCh := make(chan struct{}, 3)
	errCh := make(chan error, 1)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	var reportErrOnce sync.Once
	reportErr := func(err error) {
		if err == nil {
			return
		}
		reportErrOnce.Do(func() {
			select {
			case errCh <- err:
			default:
			}
			cancel()
		})
	}

	go receiveFromStream(streamCtx, stream, requestCh, reportErr, doneCh)
	go processRequests(streamCtx, requestCh, responseCh, httpClient, cfg.LocalPort, workerDone)
	go sendToStream(streamCtx, stream, responseCh, workerDone, reportErr, doneCh)

	finished := 0
	for finished < 2 {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case <-doneCh:
			finished++
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			return ctx.Err()
		}
	}

	return nil
}

func connectOrCreateTunnel(ctx context.Context, cfg ConnectConfig) (*api.ConnectToTunnelResponse, error) {
	endpoint, err := url.JoinPath(strings.TrimRight(cfg.ServerURL, "/"), "/api/connect")
	if err != nil {
		return nil, fmt.Errorf("failed to build connect endpoint: %w", err)
	}

	body := api.ConnectToTunnelRequest{TunnelName: cfg.Name}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to encode connect request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create connect request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connect request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading connect response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("connect request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	connectResp := &api.ConnectToTunnelResponse{}
	if err := json.Unmarshal(respBody, connectResp); err != nil {
		return nil, fmt.Errorf("failed to decode connect response: %w", err)
	}

	if strings.TrimSpace(connectResp.TunnelIp) == "" || connectResp.Port <= 0 {
		return nil, fmt.Errorf("connect response missing tunnel address")
	}

	return connectResp, nil
}

func receiveFromStream(
	ctx context.Context,
	stream tunnelv1.TunnelService_OpenTunnelClient,
	requestCh chan<- *tunnelv1.HttpRequest,
	reportErr func(error),
	doneCh chan<- struct{},
) {
	defer close(requestCh)
	defer func() { doneCh <- struct{}{} }()

	for {
		httpRequest, err := stream.Recv()
		if err != nil {
			if shouldIgnoreStreamErr(ctx, err) {
				return
			}
			reportErr(fmt.Errorf("failed receiving from stream: %w", err))
			return
		}

		select {
		case <-ctx.Done():
			return
		case requestCh <- httpRequest:
		}
	}
}

func processRequests(
	ctx context.Context,
	requestCh <-chan *tunnelv1.HttpRequest,
	responseCh chan<- *tunnelv1.HttpResponse,
	httpClient *http.Client,
	localPort int,
	workerDone chan<- struct{},
) {
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			close(responseCh)
			close(workerDone)
			return
		case req, ok := <-requestCh:
			if !ok {
				wg.Wait()
				close(responseCh)
				close(workerDone)
				return
			}

			if !req.GetAssignedToRespond() {
				continue
			}

			wg.Add(1)
			go func(httpReq *tunnelv1.HttpRequest) {
				defer wg.Done()

				httpResp := forwardToLocalHTTP(ctx, httpClient, httpReq, localPort)
				select {
				case <-ctx.Done():
				case responseCh <- httpResp:
				}
			}(req)
		}
	}
}

func sendToStream(
	ctx context.Context,
	stream tunnelv1.TunnelService_OpenTunnelClient,
	responseCh <-chan *tunnelv1.HttpResponse,
	workerDone <-chan struct{},
	reportErr func(error),
	doneCh chan<- struct{},
) {
	defer func() { doneCh <- struct{}{} }()

	for {
		select {
		case <-ctx.Done():
			return
		case <-workerDone:
			for response := range responseCh {
				if err := stream.Send(response); err != nil {
					if shouldIgnoreStreamErr(ctx, err) {
						return
					}
					reportErr(fmt.Errorf("failed sending to stream: %w", err))
					return
				}
			}
			return
		case response, ok := <-responseCh:
			if !ok {
				return
			}
			if err := stream.Send(response); err != nil {
				if shouldIgnoreStreamErr(ctx, err) {
					return
				}
				reportErr(fmt.Errorf("failed sending to stream: %w", err))
				return
			}
		}
	}
}

func forwardToLocalHTTP(ctx context.Context, httpClient *http.Client, grpcRequest *tunnelv1.HttpRequest, localPort int) *tunnelv1.HttpResponse {
	requestURL := buildLocalURL(localPort, grpcRequest.GetPath(), grpcRequest.GetRawQueryParams())

	method := grpcRequest.GetRequestMethod()
	if method == "" {
		method = http.MethodGet
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(grpcRequest.GetRequestBody()))
	if err != nil {
		return makeErrorResponse(grpcRequest.GetRequestId(), http.StatusBadGateway, fmt.Sprintf("failed to build local request: %v", err))
	}

	for _, h := range grpcRequest.GetHeaders() {
		key := strings.TrimSpace(h.GetKey())
		if key == "" {
			continue
		}
		req.Header.Add(key, h.GetValue())
	}

	localResp, err := httpClient.Do(req)
	if err != nil {
		return makeErrorResponse(grpcRequest.GetRequestId(), http.StatusBadGateway, fmt.Sprintf("local request failed: %v", err))
	}
	defer localResp.Body.Close()

	respBody, err := io.ReadAll(localResp.Body)
	if err != nil {
		return makeErrorResponse(grpcRequest.GetRequestId(), http.StatusBadGateway, fmt.Sprintf("failed reading local response: %v", err))
	}

	headers := make([]*tunnelv1.Header, 0, len(localResp.Header))
	for key, values := range localResp.Header {
		for _, value := range values {
			headers = append(headers, &tunnelv1.Header{Key: key, Value: value})
		}
	}

	return &tunnelv1.HttpResponse{
		RequestId:    grpcRequest.GetRequestId(),
		StatusCode:   int32(localResp.StatusCode),
		ResponseBody: respBody,
		Headers:      headers,
	}
}

func buildLocalURL(localPort int, path string, rawQuery string) string {
	resolvedPath := path
	if resolvedPath == "" {
		resolvedPath = "/"
	}
	base := "http://localhost:" + strconv.Itoa(localPort) + resolvedPath
	if rawQuery == "" {
		return base
	}
	return base + "?" + rawQuery
}

func makeErrorResponse(requestID string, statusCode int, message string) *tunnelv1.HttpResponse {
	return &tunnelv1.HttpResponse{
		RequestId:    requestID,
		StatusCode:   int32(statusCode),
		ResponseBody: []byte(message),
		Headers: []*tunnelv1.Header{
			{Key: "Content-Type", Value: "text/plain; charset=utf-8"},
		},
	}
}

func shouldIgnoreStreamErr(ctx context.Context, err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	if ctx.Err() != nil {
		return true
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.Canceled, codes.Unavailable:
		return true
	default:
		return false
	}
}
