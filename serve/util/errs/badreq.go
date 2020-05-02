package errs

import "errors"

// BadRequestError is used to indicate that the caller issued an
// invalid or malformed request.
type BadRequestError struct {
	error
}

// NewBadRequestError creates a BadRequestError with the specified message.
func NewBadRequestError(msg string) BadRequestError {
	return BadRequestError{errors.New(msg)}
}
