package expect

import "time"

type Option func(opts *Options) error

type Options struct {
	timeout time.Duration
	delay   time.Duration
}

func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) error {
		opts.timeout = timeout
		return nil
	}
}

func WithDelay(delay time.Duration) Option {
	return func(opts *Options) error {
		opts.delay = delay
		return nil
	}
}
