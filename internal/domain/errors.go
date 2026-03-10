package domain

import "errors"

var (
	ErrNotFound             = errors.New("user not found")
	ErrEmailTaken           = errors.New("email is already taken")
	ErrPreconditionRequired = errors.New("version token is required for this operation")
	ErrPreconditionFailed   = errors.New("resource version mismatch; re-fetch and retry")
	ErrInvalidETag          = errors.New("invalid version token format")
	ErrInvalidCursor        = errors.New("invalid pagination cursor")
)
