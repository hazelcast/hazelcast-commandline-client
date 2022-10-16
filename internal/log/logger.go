package log

type Logger interface {
	Info(format string, args ...any)
	Debug(func() string)
}
