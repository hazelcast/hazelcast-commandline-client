package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/simplehistory"

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
	rl            *readline.Editor
	endLineFn     EndLineFn
	textFn        TextFn
	prompt1       string
	prompt2       string
	historyPath   string
	stderr        io.Writer
	stdout        io.Writer
	commentPrefix string
	history       *simplehistory.Container
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
	history := simplehistory.New()
	w := colorable.NewColorableStdout()
	/*
		// TODO:
		var w io.Writer
		if color.NoColor {
			w = colorable.NewNonColorable(stdout)
		} else {
			w = colorable.NewColorableStdout()
		}
	*/
	rl := readline.Editor{
		Prompt: func() (int, error) {
			return fmt.Fprint(stdout, prompt1)
		},
		Writer:         w,
		History:        history,
		HistoryCycling: true,
		Coloring:       &SQLColoring{},
	}
	return &Shell{
		rl:            &rl,
		endLineFn:     endLineFn,
		textFn:        textFn,
		prompt1:       prompt1,
		prompt2:       prompt2,
		historyPath:   historyPath,
		stderr:        stderr,
		stdout:        w,
		commentPrefix: "",
		history:       history,
	}, nil
}

func (sh *Shell) Close() error {
	//return sh.rl.Close()
	return nil
}

func (sh *Shell) SetCommentPrefix(pfx string) {
	sh.commentPrefix = pfx
}

func (sh *Shell) Start(ctx context.Context) error {
	for {
		text, err := sh.readTextReadline()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			I2(fmt.Fprintf(sh.stderr, "%s\n", err.Error()))
		}
		if text == "" {
			continue
		}
		sh.history.Add(text)
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
		if err := sh.textFn(ctx, text); err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				I2(fmt.Fprintf(sh.stderr, color.RedString("Error: %s\n", err.Error())))
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
		sh.rl.Prompt = func() (int, error) {
			return fmt.Fprint(sh.stdout, prompt)
		}
		p, err := sh.rl.ReadLine(context.Background())
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

type SQLColoring struct {
	text []rune
}

func (c *SQLColoring) Init() int {
	c.text = nil
	return readline.DefaultForeGroundColor
}

func (c *SQLColoring) Next(r rune) int {
	return readline.DefaultForeGroundColor
}
