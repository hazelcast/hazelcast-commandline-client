package internal

import (
	"fmt"
	"io"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type Logger struct {
	w io.WriteCloser
}

func NewLogger(w io.WriteCloser) *Logger {
	return &Logger{w: w}
}

func (l Logger) Info(format string, args ...any) {
	I2(fmt.Fprintf(l.w, format, args...))
	I2(fmt.Fprintln(l.w))
}
