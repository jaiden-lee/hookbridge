package utils

import "errors"

var (
	ErrInvalidJWT = errors.New("the jwt/access token is invalid/expired")
	ErrExpiredJWT = errors.New("JWT expired")
)
