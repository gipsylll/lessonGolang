package domain

import "errors"

var (
	ErrNotFound             = errors.New("user not found")
	ErrEmailTaken           = errors.New("email is already taken")
	ErrPreconditionRequired = errors.New("version token is required for this operation")
	ErrPreconditionFailed   = errors.New("resource version mismatch; re-fetch and retry")
	ErrInvalidETag          = errors.New("invalid version token format")
	ErrInvalidCursor        = errors.New("invalid pagination cursor")
	ErrNoFieldsToUpdate     = errors.New("at least one field must be provided")
)

type FieldError struct {
	Field   string
	Message string
}

type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string { return "validation failed" }
