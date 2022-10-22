package log

type Logger interface {
	Error(err error)
	Warn(format string, args ...any)
	Info(format string, args ...any)
	Debug(func() string)
	Trace(f func() string)
}
