package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrStreamNotFound     = errors.New("stream not found")
	ErrUnauthorized       = errors.New("unauthorized")
)
