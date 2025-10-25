package db

import "errors"

var (
	ErrProjectNotFound           = errors.New("project not found")
	ErrInvalidPassword           = errors.New("invalid password")
	ErrPasswordTooLong           = errors.New("password too long, must be less than 72 chars")
	ErrProjectAlreadyExists      = errors.New("project id generated already exists. try again later")
	ErrPasswordTooShort          = errors.New("password too short. must be >= 6 chars")
	ErrPasswordSpecialCharacters = errors.New("no special characters in passwords are allowed")
	ErrNoProjectNameSpecified    = errors.New("no project name was specified. names must be at least 1 character")
)
