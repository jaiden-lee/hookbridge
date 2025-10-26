package utils

import "errors"

var (
	ErrRefreshTokenFail = errors.New("failed to refresh token, you have been signed out")
	ErrNetworkError     = errors.New("network error has occurred. please try again later")
	ErrUnexpected       = errors.New("an unexpeced error has occurred, please try again later")
)
