package clc

import "io"

type NopWriteCloser struct {
	W io.Writer
}

func (nc NopWriteCloser) Write(p []byte) (n int, err error) {
	return nc.W.Write(p)
}

func (nc NopWriteCloser) Close() error {
	return nil
}
