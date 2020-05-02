package errs

import "errors"

// UnauthorizedError is used to indicate that the caller is not allowed
// to issue the attempted request.
type UnauthorizedError struct {
	error
}

// NewUnauthorizedError creates an UnauthorizedError with the specified message.
func NewUnauthorizedError(msg string) UnauthorizedError {
	return UnauthorizedError{errors.New(msg)}
}
