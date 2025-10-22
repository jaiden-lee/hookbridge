package utils

import (
	"os"
	"time"

	"log"

	"context"
	"strings"

	"github.com/MicahParks/keyfunc"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/workos/workos-go/v5/pkg/usermanagement"
)

type UserData struct {
	UserID string
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

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidJWT
	}

	return &UserData{UserID: userID}, nil
}

func (s *AuthServiceStruct) fetchJWKs() {
	jwksURL, err := usermanagement.GetJWKSURL(s.ClientID)
	if err != nil {
		panic("failed ot fetch JWKs")
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
