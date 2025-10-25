package cmd

import (
	"net/http"
	"net/url"

	"strings"

	"encoding/json"
	"errors"

	"fmt"

	"time"

	"context"

	"github.com/jaiden-lee/hookbridge/internal/cli/config"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "sign in or create a new account",
	Long:  `sign in or create a new account`,
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := config.LoadUserConfig()
		if err != nil {
			return err
		}

		if config.IsLoggedIn(user) {
			fmt.Println("\nYou are already logged in as:", user.Email)
			fmt.Println()
			return nil
		}

		deviceResponse, err := getSignInURLAndCode()
		if err != nil {
			return err
		}

		fmt.Println("\nNOTE: you only need to sign in if you are creating a new project")
		fmt.Println("\nFollow these steps to log in:")
		fmt.Println("	1. Visit this url:", deviceResponse.VerificationURI)
		fmt.Println("	2. Enter this code:", deviceResponse.UserCode)

		tokenResponse, err := pollForToken(deviceResponse)
		if err != nil {
			return err
		}

		// save accesstoken, refreshtoken, and user email
		userConfig := config.UserConfig{
			AccessToken:  tokenResponse.AccessToken,
			RefreshToken: tokenResponse.RefreshToken,
			Email:        tokenResponse.User.Email,
		}
		config.SaveUserConfig(&userConfig)

		fmt.Println("\n✅ Login successful!")
		fmt.Println("\n Logged in as:", tokenResponse.User.Email)
		fmt.Println()

		return nil
	},
}

type DeviceResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	Interval                int    `json:"interval"`
	ExpiresIn               int    `json:"expires_in"`
}

type TokenResponse struct {
	User                 User   `json:"user"`
	AccessToken          string `json:"access_token"`
	RefreshToken         string `json:"refresh_token"`
	AuthenticationMethod string `json:"authentication_method"`
	Error                string `json:"error"`
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func getSignInURLAndCode() (*DeviceResponse, error) {
	form := url.Values{}
	form.Set("client_id", config.WorkOSClientID)

	response, err := http.Post(
		"https://api.workos.com/user_management/authorize/device",
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected error occured, please try again later")
	}

	var result DeviceResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func pollForToken(deviceResponse *DeviceResponse) (*TokenResponse, error) {
	timeout := time.Duration(2*deviceResponse.ExpiresIn) * time.Second
	pollingInterval := time.Duration(deviceResponse.Interval) * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("authorization failed due to timeout, please try again")
		default:
			// poll

			requestBody := url.Values{}
			requestBody.Add("client_id", config.WorkOSClientID)
			requestBody.Add("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
			requestBody.Add("device_code", deviceResponse.DeviceCode)

			response, err := http.Post(
				"https://api.workos.com/user_management/authenticate",
				"application/x-www-form-urlencoded",
				strings.NewReader(requestBody.Encode()),
			)

			if err != nil {
				return nil, err
			}

			defer response.Body.Close()

			bodyReader := json.NewDecoder(response.Body)
			var body TokenResponse
			err = bodyReader.Decode(&body)

			if err != nil {
				return nil, err
			}

			if body.Error == "authorization_pending" {
				// do nothing, keep going
			} else if body.Error == "invalid_grant" || body.Error == "access_denied" || body.Error == "expired_token" {
				// stop polling
				return nil, errors.New("authorized failed, please try again later")
			} else if body.Error == "slow_down" {
				// increase polling interval
				timeout *= 2 // double polling interval
			} else {
				// no error
				return &body, nil
			}

			time.Sleep(pollingInterval)
		}
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
