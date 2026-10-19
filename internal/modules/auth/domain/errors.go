package domain

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidToken = errors.New("invalid token")

	ErrInvalidCredentials          = errors.New("invalid credentials")
	ErrAccountLocked               = errors.New("account locked")
	ErrAccountDisabled             = errors.New("account disabled")
	ErrIdentityProviderUnavailable = errors.New("identity provider unavailable")
)
