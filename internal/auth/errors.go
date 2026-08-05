package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
	ErrInvalidGoogleToken = errors.New("invalid google id token")
	ErrUserNotFound       = errors.New("user not found")
)

var (
	ErrEmailAlreadyUsed        = errors.New("email already in use")
	ErrInvalidOTP              = errors.New("invalid or expired OTP")
	ErrNoActiveVerification    = errors.New("no active verification request")
	ErrPasswordSameAsCurrent   = errors.New("new password must be different from current")
	ErrCurrentPasswordRequired = errors.New("current password is required")
)
