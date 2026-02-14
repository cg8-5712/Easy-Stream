package repository

import "errors"

var (
	ErrStreamNotFound    = errors.New("stream not found")
	ErrShareLinkNotFound = errors.New("share link not found")
	ErrUserNotFound      = errors.New("user not found")
)
