package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/fatih/color"
	"github.com/gohxs/readline"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

const (
	envStyler    = "CLC_EXPERIMENTAL_STYLER"
	envFormatter = "CLC_EXPERIMENTAL_FORMATTER"
)

var ErrExit = errors.New("exit")

type EndLineFn func(line string, multiline bool) (string, bool)

type TextFn func(ctx context.Context, text string) error

type Shell struct {
	rl            *readline.Instance
	endLineFn     EndLineFn
	textFn        TextFn
	prompt1       string
	prompt2       string
	historyPath   string
	stderr        io.Writer
	stdout        io.Writer
	commentPrefix string
}

func New(prompt1, prompt2, historyPath, lexer string, stdout, stderr io.Writer, endLineFn EndLineFn, textFn TextFn) (*Shell, error) {
	var styler string
	if !color.NoColor {
		styler = os.Getenv(envStyler)
		if styler == "" {
			styler = "clc-default"
		}
	}
	formatter := os.Getenv(envFormatter)
	if formatter == "" || !strings.HasPrefix(formatter, "terminal") {
		formatter = "terminal"
	}
	cfg := &readline.Config{
		Prompt:          prompt1,
		HistoryFile:     historyPath,
		InterruptPrompt: "^C",
		EOFPrompt:       `\exit`,
		Output: func(input string) string {
			if lexer == "" || styler == "" {
				return input
			}
			buf := bytes.NewBuffer([]byte{})
			err := quick.Highlight(buf, input, lexer, formatter, styler)
			if err != nil {
				log.Fatal(err)
			}
			return buf.String()
		},
		HistorySearchFold: true,
		Stdout:            stdout,
		Stderr:            stderr,
	}
	if historyPath != "" {
		cfg.HistoryFile = historyPath
	}
	rl, err := readline.NewEx(cfg)
	if err != nil {
		return nil, err
	}
	return &Shell{
		rl:            rl,
		endLineFn:     endLineFn,
		textFn:        textFn,
		prompt1:       prompt1,
		prompt2:       prompt2,
		historyPath:   historyPath,
		stderr:        stderr,
		stdout:        stdout,
		commentPrefix: "",
	}, nil
}

func (sh *Shell) Close() error {
	return sh.rl.Close()
}

func (sh *Shell) SetCommentPrefix(pfx string) {
	sh.commentPrefix = pfx
}

func (sh *Shell) Start(ctx context.Context) error {
	for {
		text, err := sh.readTextReadline()
		if err == readline.ErrInterrupt || err == io.EOF {
			return nil
		}
		if err != nil {
			I2(fmt.Fprintf(sh.stderr, "%s\n", err.Error()))
		}
		if text == "" {
			continue
		}
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
		if err := sh.textFn(ctx, text); err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				I2(fmt.Fprintf(sh.stderr, "Error: %s\n", err.Error()))
			}
		}
		stop()
	}
}

func (sh *Shell) readTextReadline() (string, error) {
	// NOTE: when this implementation is changed,
	// clc/shell/oneshot_shell.go:readTextBasic should also change!
	prompt := sh.prompt1
	multiline := false
	var sb strings.Builder
	for {
		sh.rl.SetPrompt(prompt)
		p, err := sh.rl.Readline()
		if err != nil {
			return "", err
		}
		if !multiline {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if sh.commentPrefix != "" && strings.HasPrefix(p, sh.commentPrefix) {
				continue
			}
		}
		text, end := sh.endLineFn(p, multiline)
		sb.WriteString(text)
		multiline = !end
		if end {
			break
		}
		prompt = sh.prompt2
	}
	text := sb.String()
	return text, nil
}
