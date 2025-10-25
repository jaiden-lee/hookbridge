package handlers

import (
	"errors"
)

var (
	ErrInvalidUserInContext = errors.New("either no user was passed in the context, or the user is invalid")
)
