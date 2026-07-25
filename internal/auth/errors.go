package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
	ErrInvalidGoogleToken = errors.New("invalid google id token")
	ErrUserNotFound       = errors.New("user not found")
)
