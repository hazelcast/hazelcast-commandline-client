package it

import (
	"bytes"
	"fmt"
	"io"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
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
	l.log("INFO", fmt.Sprintf(format, args...))
	I2(fmt.Fprintf(l.buf, format, args...))
}

func (l Logger) Error(err error) {
	l.log("ERROR", fmt.Sprint(err.Error()))
	I2(fmt.Fprint(l.buf, err.Error()))
}

func (l Logger) Warn(format string, args ...any) {
	l.log("WARN", fmt.Sprintf(format, args...))
	I2(fmt.Fprintf(l.buf, format, args...))
}

func (l Logger) Debug(f func() string) {
	text := f()
	l.log("DEBUG", fmt.Sprint(text))
	I2(fmt.Fprint(l.buf, text))
}

func (l Logger) Debugf(format string, args ...any) {
	l.log("DEBUG", fmt.Sprintf(format, args...))
	I2(fmt.Fprintf(l.buf, format, args...))
}

func (l Logger) Trace(f func() string) {
	text := f()
	l.log("TRACE", fmt.Sprint(text))
	I2(fmt.Fprint(l.buf, text))
}

func (l Logger) log(level, text string) {
	I2(fmt.Fprintf(l.buf, "%s %s", level, text))
}
