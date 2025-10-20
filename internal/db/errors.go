package db

import "errors"

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrPasswordTooLong      = errors.New("password too long, must be less than 72 chars")
	ErrProjectAlreadyExists = errors.New("project id generated already exists. try again later")
)
