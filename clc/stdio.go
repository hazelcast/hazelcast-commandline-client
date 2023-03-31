package clc

import (
	"io"
	"os"
)

type IO struct {
	Stdin  io.Reader
	Stderr io.Writer
	Stdout io.Writer
}

func StdIO() IO {
	return IO{
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
		Stdout: os.Stdout,
	}
}
