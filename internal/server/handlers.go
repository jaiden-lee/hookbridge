package server

import (
	"context"
	"encoding/json"
	"hookbridge/internal/api"
	"io"
	"strings"
	"time"

	"log"
	"regexp"

	"hookbridge/gen/tunnelv1"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type connectHandlersStruct struct{}
type tunnelHandlersStruct struct{}

var connectHandlers = connectHandlersStruct{}
var tunnelHandlers = tunnelHandlersStruct{}

func (_ *connectHandlersStruct) connectOrCreateTunnel(c *gin.Context) {
	var requestBody api.ConnectToTunnelRequest

	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil {
		log.Println("json request body was invalid")
		log.Println(err)
		c.AbortWithStatusJSON(400, gin.H{
			"error": "invalid json request body",
		})
		return
	}

	if requestBody.TunnelName == "" {
		log.Println("tunnel_name parameter was empty")
		c.AbortWithStatusJSON(400, gin.H{
			"error": "tunnel_name parameter is empty",
		})
		return
	}

	if !isValidTunnelName(requestBody.TunnelName) {
		log.Println("tunnel name was invalid; no special characters allowed")
		c.AbortWithStatusJSON(400, gin.H{
			"error": "invalid tunnel name, no spaces or special characters besides _ and - are allowed",
		})
		return
	}

	tunnelName := strings.ToLower(requestBody.TunnelName)

	log.Printf("Connecting to tunnel <%s>\n", tunnelName)

	port, tunnelExists := serverState.activeTunnels[tunnelName]

	if !tunnelExists {
		log.Println("Tunnel doesn't exist. Creating tunnel...")

		// this is created beforehand, since tunnel dials server immediately; so needs to exist in map beforehand
		serverState.tunnelRequestChannels[tunnelName] = make(chan *tunnelv1.HttpRequest)
		// serverState.tunnelResponseCHannels[tunnelName] = make(chan *tunnelv1.HttpResponse)
		port, err = startTunnel(tunnelName)
		serverState.activeTunnels[tunnelName] = port

		if err != nil {
			log.Println("Failed to start tunnel docker container")
			log.Println(err)

			cleanupTunnel(tunnelName)

			c.AbortWithStatusJSON(500, gin.H{
				"error": "internal server error, failed to start tunnel docker container",
			})
			return
		}

	}

	log.Printf("Tunnel is ready. Returning port number %d to client...\n", port)

	c.JSON(200, api.ConnectToTunnelResponse{
		TunnelIp: serverState.serverIp,
		Port:     port,
	})
}

func (_ *tunnelHandlersStruct) handleHttpRequestForTunnel(c *gin.Context) {
	tunnelName := c.Param("tunnel_name")
	proxyPath := c.Param("proxyPath") // includes leading /
	queryParams := c.Request.URL.RawQuery

	requestChannel, requestExists := serverState.tunnelRequestChannels[tunnelName]

	if !requestExists {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "This tunnel doesn't exist",
		})
		return
	}

	// Create the protobuf HttpRequest
	requestId, err := uuid.NewUUID()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": "Failed to generate ID for the request. Try again later.",
		})
		return
	}
	requestIdStr := requestId.String()

	responseChannel := make(chan *tunnelv1.HttpResponse, 1) // nonblocking on sender side
	serverState.tunnelResponseCHannels[requestIdStr] = responseChannel
	defer cleanupRequest(requestIdStr)

	requestMethod := c.Request.Method
	if requestMethod == "" { // empty string also means GET
		requestMethod = "GET"
	}

	requestBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": "Failed to read request body. Try again later.",
		})
		return
	}

	requestHeaders := []*tunnelv1.Header{}
	for key, value := range c.Request.Header {
		header := tunnelv1.Header{
			Key:   key,
			Value: strings.Join(value, ", "),
		}
		requestHeaders = append(requestHeaders, &header)
	}

	httpRequest := &tunnelv1.HttpRequest{
		RequestId:         requestIdStr,
		AssignedToRespond: false, // this gets chosen later
		RequestMethod:     requestMethod,
		Path:              proxyPath,
		RawQueryParams:    queryParams, // no ? included
		RequestBody:       requestBody,
		Headers:           requestHeaders,
	}

	// Set timeout context
	timeoutContext, cancel := context.WithTimeout(c.Request.Context(), time.Duration(serverState.requestTimeout)*time.Second)
	defer cancel() // still need cancel to free resources

	// Forward HttpRequest
	select {
	case <-timeoutContext.Done():
		c.AbortWithStatusJSON(408, gin.H{
			"error": "Request timeout or request was cancelled",
		})
		return
	case requestChannel <- httpRequest:
	}

	// Listen for HttpResponse
	select {
	case <-timeoutContext.Done():
		c.AbortWithStatusJSON(408, gin.H{
			"error": "Request timeout or request was cancelled",
		})
		return
	case httpResponse, ok := <-responseChannel:
		if !ok {
			c.AbortWithStatusJSON(400, gin.H{
				"error": "Tunnel was closed while this request was processing.",
			})
			return
		}
		// return httpresponse
		contentType := "application/octet-stream"
		for _, h := range httpResponse.Headers {
			k := h.Key
			v := h.Value

			// Skip hop-by-hop headers
			if isHopByHopHeader(k) {
				continue
			}

			if strings.EqualFold(k, "Content-Type") {
				contentType = v
				c.Writer.Header().Set(k, v) // don't Add
				continue
			}

			if strings.EqualFold(k, "Set-Cookie") {
				c.Writer.Header().Add(k, v) // can repeat
				continue
			}

			// Most headers should be Set (single value)
			c.Writer.Header().Set(k, v)
		}

		c.Data(int(httpResponse.StatusCode), contentType, httpResponse.ResponseBody)
		return
	}
}

func isHopByHopHeader(k string) bool {
	switch {
	case strings.EqualFold(k, "Connection"),
		strings.EqualFold(k, "Keep-Alive"),
		strings.EqualFold(k, "Proxy-Authenticate"),
		strings.EqualFold(k, "Proxy-Authorization"),
		strings.EqualFold(k, "TE"),
		strings.EqualFold(k, "Trailer"),
		strings.EqualFold(k, "Transfer-Encoding"),
		strings.EqualFold(k, "Upgrade"):
		return true
	default:
		return false
	}
}

func isValidTunnelName(tunnelName string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9_-]+$`) // _, -, abc..., 123...    no space or special chars
	return re.MatchString(tunnelName)
}

func cleanupTunnel(tunnelName string) {
	_, ok := serverState.tunnelRequestChannels[tunnelName]
	if ok {
		// close(requestChan) let be garbage collected
		delete(serverState.tunnelRequestChannels, tunnelName)
	}

	// _, ok = serverState.tunnelResponseCHannels[tunnelName]
	// if ok {
	// 	// close(responseChan)
	// 	// don't close responseChan; let it be garbage collected
	// 	delete(serverState.tunnelResponseCHannels, tunnelName)
	// }

	if _, ok = serverState.activeTunnels[tunnelName]; ok {
		delete(serverState.activeTunnels, tunnelName) // port number
	}
}

func cleanupRequest(requestId string) {
	_, ok := serverState.tunnelResponseCHannels[requestId]
	if ok {
		// close(responseChan)
		// don't close responseChan; let it be garbage collected
		delete(serverState.tunnelResponseCHannels, requestId)
	}
}
