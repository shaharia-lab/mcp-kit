package google

import "errors"

var (
	ErrNoTokenAvailable = errors.New("no token available")
	ErrInvalidState     = errors.New("invalid state parameter")
	ErrMissingCode      = errors.New("authorization code is missing")
)
