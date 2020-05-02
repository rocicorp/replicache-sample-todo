package errs

type SequenceError struct {
	error
}

func NewSequenceError(err error) SequenceError {
	return SequenceError{err}
}
