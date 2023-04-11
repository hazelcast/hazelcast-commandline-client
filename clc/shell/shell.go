package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/gohxs/readline"
	"github.com/mattn/go-colorable"
	ny "github.com/nyaosorg/go-readline-ny"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

const (
	CmdPrefix    = `\`
	envStyler    = "CLC_EXPERIMENTAL_STYLER"
	envFormatter = "CLC_EXPERIMENTAL_FORMATTER"
	EnvReadline  = "CLC_EXPERIMENTAL_READLINE"
)

var ErrExit = errors.New("exit")

type EndLineFn func(line string, multiline bool) (string, bool)

type TextFn func(ctx context.Context, stdout io.Writer, text string) error

type Shell struct {
	lr            LineReader
	endLineFn     EndLineFn
	textFn        TextFn
	prompt1       string
	prompt2       string
	historyPath   string
	stderr        io.Writer
	stdout        io.Writer
	stdin         io.Reader
	commentPrefix string
}

func New(prompt1, prompt2, historyPath string, stdout, stderr io.Writer, stdin io.Reader, endLineFn EndLineFn, textFn TextFn) (*Shell, error) {
	stdout, stderr = fixStdoutStderr(stdout, stderr)
	sh := &Shell{
		endLineFn:     endLineFn,
		textFn:        textFn,
		prompt1:       prompt1,
		prompt2:       prompt2,
		historyPath:   historyPath,
		stderr:        stderr,
		stdout:        stdout,
		stdin:         stdin,
		commentPrefix: "",
	}
	rl := os.Getenv(EnvReadline)
	if rl == "" && runtime.GOOS == "windows" {
		// ny is default on Windows
		rl = "ny"
	}
	if rl == "ny" {
		if err := sh.createNyLineReader(prompt1); err != nil {
			return nil, err
		}
	} else if err := sh.createGohxsLineReader(prompt1); err != nil {
		return nil, err
	}
	return sh, nil
}

func (sh *Shell) Close() error {
	return sh.lr.Close()
}

func (sh *Shell) SetCommentPrefix(pfx string) {
	sh.commentPrefix = pfx
}

func (sh *Shell) Start(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()
	for {
		text, err := sh.readTextReadline()
		if err == io.EOF {
			return nil
		}
		if err == readline.ErrInterrupt || err != nil && err.Error() == "^C" {
			I2(fmt.Fprintf(sh.stderr, color.RedString("Press Ctrl+D or type \\exit to exit.\n")))
			continue
		}
		if err != nil {
			I2(fmt.Fprintf(sh.stderr, color.RedString("Error: %s\n", err.Error())))
		}
		if text == "" {
			continue
		}
		sh.lr.AddToHistory(text)
		if err := sh.textFn(ctx, sh.stdout, text); err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}
		}
	}
}

func (sh *Shell) readTextReadline() (string, error) {
	// NOTE: when this implementation is changed,
	// clc/shell/oneshot_shell.go:readTextBasic should also change!
	prompt := sh.prompt1
	multiline := false
	var sb strings.Builder
	for {
		sh.lr.SetPrompt(prompt)
		p, err := sh.lr.ReadLine(context.Background())
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
	return ny.DefaultForeGroundColor
}

func (c *SQLColoring) Next(_ rune) int {
	return ny.DefaultForeGroundColor
}

func isStdout(w io.Writer) bool {
	if wc, ok := w.(clc.NopWriteCloser); ok {
		return wc.W == os.Stdout
	}
	return false
}

func isStderr(w io.Writer) bool {
	if wc, ok := w.(clc.NopWriteCloser); ok {
		return wc.W == os.Stderr
	}
	return false
}

// fixStdoutStderr fixes stdout and stderr on Windows, so escape codes are not printed.
func fixStdoutStderr(stdout, stderr io.Writer) (io.Writer, io.Writer) {
	if isStdout(stdout) {
		if color.NoColor {
			stdout = colorable.NewNonColorable(stdout)
		} else {
			// colorable.NewNonColorable doesn't work well on non-Windows
			stdout = colorable.NewColorableStdout()
		}
	}
	if isStderr(stderr) {
		if color.NoColor {
			stderr = colorable.NewNonColorable(stderr)
		} else {
			// colorable.NewNonColorable doesn't work well on non-Windows
			stderr = colorable.NewColorableStderr()
		}
	}
	return stdout, stderr
}
