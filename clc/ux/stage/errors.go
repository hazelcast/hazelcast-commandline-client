package stage

type ignoreError struct {
	Err error
}

func (se ignoreError) Unwrap() error {
	return se.Err
}

func (se ignoreError) Error() string {
	return se.Err.Error()
}

func IgnoreError(wrappedErr error) error {
	return ignoreError{Err: wrappedErr}
}
