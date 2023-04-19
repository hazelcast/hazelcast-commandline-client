package clc

import (
	"io"
)

type NopWriteCloser struct {
	W io.Writer
}

func (nc NopWriteCloser) Write(p []byte) (n int, err error) {
	return nc.W.Write(p)
}

func (nc NopWriteCloser) Close() error {
	return nil
}

type Spinner interface {
	Start() error
	SetText(text string)
	// SetProgress sets the progress of an operation.
	// progress should be in the range of [0, 1].
	SetProgress(progress float32)
}
