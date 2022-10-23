package shell

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/peterh/liner"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type EndLineFn func(line string) bool

type TextFn func(ctx context.Context, text string) error

type Shell struct {
	state       *liner.State
	endLineFn   EndLineFn
	textFn      TextFn
	prompt1     string
	prompt2     string
	historyPath string
	stderr      io.Writer
	stdout      io.Writer
}

func New(prompt1, prompt2, historyPath string, stdout, stderr io.Writer, endLineFn EndLineFn, textFn TextFn) *Shell {
	state := liner.NewLiner()
	state.SetCtrlCAborts(true)
	state.SetMultiLineMode(true)
	return &Shell{
		state:       state,
		endLineFn:   endLineFn,
		textFn:      textFn,
		prompt1:     prompt1,
		prompt2:     prompt2,
		historyPath: historyPath,
		stderr:      stderr,
		stdout:      stdout,
	}
}

func (sh Shell) Close() error {
	return sh.state.Close()
}

func (sh Shell) Start(ctx context.Context) error {
	for {
		text, err := sh.readText()
		if err == liner.ErrPromptAborted || err == io.EOF {
			return nil
		}
		if text == "" {
			continue
		}
		if err != nil {
			I2(fmt.Fprintf(sh.stderr, "%s\n", err.Error()))
		}
		if err := sh.textFn(ctx, text); err != nil {
			I2(fmt.Fprintf(sh.stderr, "%s\n", err.Error()))
		}
	}
}

func (sh Shell) readText() (string, error) {
	prompt := sh.prompt1
	var sb strings.Builder
	for {
		p, err := sh.state.Prompt(prompt)
		if err != nil {
			return "", err
		}
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		sb.WriteString(p)
		sb.WriteString("\n")
		if sh.endLineFn(p) {
			break
		}
		prompt = sh.prompt2
	}
	text := sb.String()
	sh.state.AppendHistory(text)
	return text, nil
}
