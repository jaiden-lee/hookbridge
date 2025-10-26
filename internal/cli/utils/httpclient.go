package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"context"

	"github.com/google/go-querystring/query"
	"github.com/jaiden-lee/hookbridge/internal/cli/config"
	"github.com/workos/workos-go/v5/pkg/usermanagement"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func PostWithAuth(uri string, body any, user *config.UserConfig) ([]byte, error) {
	return doRequestWithAuth(http.MethodPost, uri, body, user)
}
func PutWithAuth(uri string, body any, user *config.UserConfig) ([]byte, error) {
	return doRequestWithAuth(http.MethodPut, uri, body, user)
}
func PatchWithAuth(uri string, body any, user *config.UserConfig) ([]byte, error) {
	return doRequestWithAuth(http.MethodPatch, uri, body, user)
}
func DeleteWithAuth(uri string, user *config.UserConfig) ([]byte, error) {
	return doRequestWithAuth(http.MethodDelete, uri, nil, user)
}
func GetWithAuth(uri string, params any, user *config.UserConfig) ([]byte, error) {
	values, err := query.Values(params)
	if err != nil {
		return nil, err
	}

	// attach params to base URL
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	u.RawQuery = values.Encode()

	return doRequestWithAuth(http.MethodGet, u.String(), nil, user)
}

func doRequestWithAuth(method string, uri string, requestBody any, user *config.UserConfig) ([]byte, error) {
	var bodyReader io.Reader
	if requestBody != nil {
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	request, err := http.NewRequest(
		method,
		uri,
		bodyReader,
	)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+user.AccessToken)

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// retry for new token
	if response.StatusCode == http.StatusUnauthorized {
		response.Body.Close() // close old one manually before retrying

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		refreshResponse, err := usermanagement.AuthenticateWithRefreshToken(
			ctx,
			usermanagement.AuthenticateWithRefreshTokenOpts{
				ClientID:     config.WorkOSClientID,
				RefreshToken: user.RefreshToken,
			},
		)

		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) {
				// 🌐 Network-level error
				return nil, ErrNetworkError
			}

			// Otherwise, it’s likely an API (auth) error
			// delete user config
			config.DeleteUserConfig()
			return nil, ErrRefreshTokenFail
		}

		newUserConfig := config.UserConfig{
			AccessToken:  refreshResponse.AccessToken,
			RefreshToken: refreshResponse.RefreshToken,
			Email:        user.Email,
		}

		config.SaveUserConfig(&newUserConfig)

		// retry request
		request.Header.Set("Authorization", "Bearer "+newUserConfig.AccessToken)
		response, err = http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		if response.StatusCode == http.StatusUnauthorized {
			// for some reason, unauthorized again; so just give up
			config.DeleteUserConfig()
			return nil, ErrRefreshTokenFail
		}
	}

	// non success responses
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var errResponse ErrorResponse

		err = json.NewDecoder(response.Body).Decode(&errResponse)
		if err == nil {
			// no error when decoding
			if errResponse.Error != "" {
				// error field attached in response
				return nil, errors.New(errResponse.Error)
			}
		}

		// otherwise, do generic error
		return nil, ErrUnexpected
	}

	// SUCCESSFUL RESPONSE
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
