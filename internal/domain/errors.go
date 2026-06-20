package domain

import "errors"

var (
	ErrInvalidID = errors.New("invalid id format")
	ErrNotFound = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
)