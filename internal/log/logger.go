package log

type Logger interface {
	Info(format string, args ...any)
}
