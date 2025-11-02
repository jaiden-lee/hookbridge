package utils

import (
	"os"
	"time"

	"log"

	"context"
	"strings"

	"net/http"

	"github.com/MicahParks/keyfunc"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/workos/workos-go/v5/pkg/usermanagement"
)

type UserData struct {
	UserID    string
	SessionID string
}

type AuthServiceStruct struct {
	APIKey   string
	ClientID string
	inited   bool
	JWKS     *keyfunc.JWKS
}

var AuthService = &AuthServiceStruct{inited: false}

func (s *AuthServiceStruct) Init() {
	if s.inited {
		return
	}

	apiKey := os.Getenv("WORKOS_API_KEY")
	clientID := os.Getenv("WORKOS_CLIENT_ID")

	if apiKey == "" || clientID == "" {
		panic("missing WORKOS_API_KEY or WORKOS_CLIENT_ID")
	}

	s.APIKey = apiKey
	s.ClientID = clientID
	s.inited = true

	usermanagement.SetAPIKey(s.APIKey)

	s.fetchJWKs()
}

func (s *AuthServiceStruct) VerifyJWT(tokenB64 string) (*UserData, error) {
	if !s.inited {
		panic("auth service not initialized")
	}

	token, err := jwt.Parse(tokenB64, s.JWKS.Keyfunc)
	if err != nil {
		// force refresh just in case jwks is expired
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = s.JWKS.Refresh(ctx, keyfunc.RefreshOptions{})

		token, err = jwt.Parse(tokenB64, s.JWKS.Keyfunc)
		if err != nil {
			return nil, ErrInvalidJWT
		}
	}

	if !token.Valid {
		return nil, ErrInvalidJWT
	}

	claims, ok := token.Claims.(jwt.MapClaims) // type assertion
	if !ok {
		return nil, ErrInvalidJWT
	}

	iss, ok := claims["iss"].(string) // type assertion, also ensures iss exists
	if !ok {
		return nil, ErrInvalidJWT
	}

	if !strings.HasPrefix(iss, "https://api.workos.com/") {
		return nil, ErrInvalidJWT
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, ErrInvalidJWT
	}

	if time.Now().Unix() > int64(exp) {
		return nil, ErrExpiredJWT
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidJWT
	}

	sessionID, ok := claims["sid"].(string)
	if !ok {
		return nil, ErrInvalidJWT
	}

	return &UserData{UserID: userID, SessionID: sessionID}, nil
}

func (s *AuthServiceStruct) fetchJWKs() {
	jwksURL, err := usermanagement.GetJWKSURL(s.ClientID)
	if err != nil {
		panic("failed to fetch JWKs")
	}

	jwks, err := keyfunc.Get(jwksURL.String(), keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshErrorHandler: func(err error) {
			log.Printf("JWKS refresh failed: %v", err)
		},
	})

	if err != nil {
		panic("failed to fetch JWKs")
	}

	s.JWKS = jwks
}

func (s *AuthServiceStruct) SignOutUser(sessionID string) {
	url, err := usermanagement.GetLogoutURL(usermanagement.GetLogoutURLOpts{
		SessionID: sessionID,
	})

	client := &http.Client{Timeout: 5 * time.Second}
	if err == nil {
		res, err := client.Get(url.String())
		if err == nil {
			defer res.Body.Close()
		}
		// don't actually care about result, if it fails then it fails
	}
	// doesn't return anything, just assume automatic success
}

func (s *AuthServiceStruct) ExchangeRefreshToken(refreshToken string) (*usermanagement.RefreshAuthenticationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	refreshResponse, err := usermanagement.AuthenticateWithRefreshToken(
		ctx,
		usermanagement.AuthenticateWithRefreshTokenOpts{
			ClientID:     s.ClientID,
			RefreshToken: refreshToken,
		},
	)

	if err != nil {
		return nil, err
	}

	return &refreshResponse, nil
}
