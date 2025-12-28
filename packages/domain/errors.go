// Package domain contains core business entities and value objects.
package domain

import "errors"

// Domain errors.
var (
	ErrNotFound            = errors.New("not found")
	ErrJobNotFound         = errors.New("job not found")
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrProviderNotFound    = errors.New("provider not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInvalidInput        = errors.New("invalid input")
	ErrProviderUnavailable = errors.New("provider unavailable")
)
