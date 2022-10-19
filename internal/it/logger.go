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

func (l Logger) Debug(f func() string) {
	//TODO implement me
	panic("implement me")
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

func (l Logger) log(level, text string) {
	I2(fmt.Fprintf(l.buf, "%s %s", level, text))
}
