package domain

import "errors"

var (
	ErrInvalidID          = errors.New("invalid id format")
	ErrNotFound           = errors.New("resource not found")
	ErrBadRequest         = errors.New("bad request")
	ErrConflict           = errors.New("resource already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid email or password")
)
