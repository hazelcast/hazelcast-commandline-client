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
