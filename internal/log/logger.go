package log

import hzlogger "github.com/hazelcast/hazelcast-go-client/logger"

type Logger interface {
	Error(err error)
	Warn(format string, args ...any)
	Info(format string, args ...any)
	Debug(func() string)
	Debugf(format string, args ...any)
	Trace(f func() string)
	Log(weight hzlogger.Weight, f func() string)
}

type NopLogger struct{}

func (NopLogger) Error(err error)                             {}
func (NopLogger) Warn(format string, args ...any)             {}
func (NopLogger) Info(format string, args ...any)             {}
func (NopLogger) Debug(func() string)                         {}
func (NopLogger) Debugf(format string, args ...any)           {}
func (NopLogger) Trace(f func() string)                       {}
func (NopLogger) Log(weight hzlogger.Weight, f func() string) {}
