package it

import (
	"bytes"
	"fmt"
	"io"
)

type Logger struct {
	buf *bytes.Buffer
}

func NewLogger() *Logger {
	return &Logger{buf: &bytes.Buffer{}}
}

func (l Logger) Text() string {
	b, err := io.ReadAll(l.buf)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (l Logger) Info(format string, args ...any) {
	_, _ = fmt.Fprintf(l.buf, format, args...)
}
