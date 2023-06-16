package jet

import "errors"

var (
	ErrInvalidJobID = errors.New("invalid job ID")
	ErrJobFailed    = errors.New("job failed")
	ErrJobNotFound  = errors.New("job not found")
)
