package domain

import "errors"

var (
	ErrNotFound             = errors.New("user not found")
	ErrPreconditionRequired = errors.New("If-Match header is required; fetch the resource first to get its ETag")
	ErrPreconditionFailed   = errors.New("resource was modified by another request; please re-fetch and retry")
)
