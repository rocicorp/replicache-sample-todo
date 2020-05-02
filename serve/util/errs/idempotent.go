package errs

type IdempotencyError struct {
	error
}

// NewIdempotencyError creates an IdempotencyError with the specified message.
func NewIdempotencyError(err error) IdempotencyError {
	return IdempotencyError{err}
}
