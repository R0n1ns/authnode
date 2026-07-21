package domain

import "errors"

var (
	ErrInvalidCode  = errors.New("invalid or expired verification code")
	ErrInvalidInput = errors.New("invalid request")
	ErrInvalidToken = errors.New("invalid or expired refresh token")
	ErrRateLimited  = errors.New("too many requests; retry later")
	ErrNotFound     = errors.New("resource not found")
)
