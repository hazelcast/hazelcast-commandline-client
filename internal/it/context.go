package it

import (
	"bytes"
	"io"

	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
)

type CommandContext struct {
}

func (c CommandContext) AddStringFlag(name, value string, required bool, help string) {
	//
}

type ExecContext struct {
	lg     *Logger
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	args   []string
}

func NewExecuteContext(args []string) *ExecContext {
	return &ExecContext{
		lg:     NewLogger(),
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		args:   args,
	}
}

func (ec *ExecContext) Logger() log.Logger {
	return ec.lg
}

func (ec *ExecContext) Stdout() io.Writer {
	return ec.stdout
}

func (ec *ExecContext) Stderr() io.Writer {
	return ec.stderr
}

func (ec *ExecContext) Args() []string {
	return ec.args
}

func (ec *ExecContext) StdoutText() string {
	b, err := io.ReadAll(ec.stdout)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (ec *ExecContext) StderrText() string {
	b, err := io.ReadAll(ec.stderr)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (ec *ExecContext) LoggerText() string {
	return ec.lg.Text()
}

func (ec *ExecContext) Set(name string, value any) {
}
func (ec *ExecContext) Get(name string) any {
	return nil
}
