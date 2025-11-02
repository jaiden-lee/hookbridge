package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
	"github.com/jaiden-lee/hookbridge/internal/cli/config"
	"github.com/jaiden-lee/hookbridge/pkg/api"
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

		refreshRequest := api.ExchangeRefreshTokenRequest{
			RefreshToken: user.RefreshToken,
		}
		refreshBody, err := json.Marshal(refreshRequest)
		if err != nil {
			return nil, err
		}

		refreshResponse, err := http.Post(
			config.APIBaseURL+"/api/auth/refresh",
			"application/json",
			bytes.NewBuffer(refreshBody),
		)

		if err != nil {
			return nil, ErrUnexpected
		}
		defer refreshResponse.Body.Close()

		if refreshResponse.StatusCode < 200 || refreshResponse.StatusCode >= 300 {
			// FAIL
			var errResponse ErrorResponse

			err = json.NewDecoder(refreshResponse.Body).Decode(&errResponse)
			if err == nil {
				// no error when decoding
				// error field attached in response; failed
				// config.DeleteUserConfig()
				return nil, ErrRefreshTokenFail
			}

			// otherwise, do generic error
			return nil, ErrUnexpected
		}

		var refreshResponseBody api.ExchangeRefreshTokenResponse
		err = json.NewDecoder(refreshResponse.Body).Decode(&refreshResponseBody)

		if err != nil {
			// generic error
			return nil, ErrRefreshTokenFail
		}

		newUserConfig := config.UserConfig{
			AccessToken:  refreshResponseBody.AccessToken,
			RefreshToken: refreshResponseBody.RefreshToken,
			Email:        user.Email,
		}

		config.SaveUserConfig(&newUserConfig)

		// retry request
		var retryBody io.Reader
		if requestBody != nil {
			jsonBody, _ := json.Marshal(requestBody)
			retryBody = bytes.NewBuffer(jsonBody)
		}
		retryReq, _ := http.NewRequest(method, uri, retryBody)
		retryReq.Header.Set("Content-Type", "application/json")
		retryReq.Header.Set("Authorization", "Bearer "+newUserConfig.AccessToken)

		response, err = http.DefaultClient.Do(retryReq)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		if response.StatusCode == http.StatusUnauthorized {
			// for some reason, unauthorized again; so just give up
			// config.DeleteUserConfig()
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
