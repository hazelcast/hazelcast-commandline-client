package errors

import "fmt"

type LoggableError struct {
	msg string
	err error
}

func NewLoggableError(err error, format string, a ...interface{}) LoggableError {
	return LoggableError{
		msg: fmt.Sprintf(format, a...),
		err: err,
	}
}

func (e LoggableError) Error() string {
	return e.msg
}

func (e LoggableError) VerboseError() string {
	return fmt.Sprintf("%s\nDetails: %s", e.msg, e.err)
}

func (e LoggableError) Unwrap() error {
	return e.err
}
