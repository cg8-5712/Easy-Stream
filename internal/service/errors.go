package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrStreamNotFound     = errors.New("stream not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrStreamExpired      = errors.New("stream has expired")
	ErrPrivateStream      = errors.New("private stream requires authentication")

	// 分享码相关错误
	ErrInvalidShareCode        = errors.New("invalid share code")
	ErrShareCodeMaxUsesReached = errors.New("share code max uses reached")
	ErrNotPrivateStream        = errors.New("only private streams support sharing")
	ErrStreamEnded             = errors.New("stream has ended")
	ErrNoShareCode             = errors.New("stream has no share code")

	// 分享链接相关错误
	ErrInvalidShareLink        = errors.New("invalid share link")
	ErrShareLinkMaxUsesReached = errors.New("share link max uses reached")
	ErrShareLinkNotFound       = errors.New("share link not found")
)
