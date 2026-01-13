package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrStreamNotFound     = errors.New("stream not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrStreamExpired      = errors.New("stream has expired")
	ErrPrivateStream      = errors.New("private stream requires password")
)
