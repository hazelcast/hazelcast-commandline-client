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
	"github.com/mattn/go-colorable"
	ny "github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/simplehistory"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

const (
	CmdPrefix       = `\`
	envStyler       = "CLC_EXPERIMENTAL_STYLER"
	envFormatter    = "CLC_EXPERIMENTAL_FORMATTER"
	envForceWindows = "CLC_FORCE_WINDOWS"
	envWtSession    = "WT_SESSION"
	maxErrorLines   = 5
)

var ErrExit = errors.New("exit")

type EndLineFn func(line string, multiline bool) (string, bool)

type TextFn func(ctx context.Context, text string) error

type Shell struct {
	lr            LineReader
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

func New(prompt1, prompt2, historyPath string, stdout, stderr io.Writer, endLineFn EndLineFn, textFn TextFn) (*Shell, error) {
	stdout, stderr = fixStdoutStderr(stdout, stderr)
	history := simplehistory.New()
	sh := &Shell{
		endLineFn:     endLineFn,
		textFn:        textFn,
		prompt1:       prompt1,
		prompt2:       prompt2,
		historyPath:   historyPath,
		stderr:        stderr,
		stdout:        stdout,
		commentPrefix: "",
		history:       history,
	}
	if windowsCommandPrompt() {
		if err := sh.createWindowsLineReader(prompt1); err != nil {
			return nil, err
		}
		return sh, nil
	}
	if err := sh.createUnixLineReader(prompt1); err != nil {
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
				I2(fmt.Fprintf(sh.stderr, color.RedString("Error: %s\n", trimError(err, maxErrorLines))))
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

// trimErrorString trims the string so it's at most n lines
func trimError(err error, n int) string {
	lines := strings.Split(err.Error(), "\n")
	if len(lines) > n {
		lines = append(lines[:5], "(Rest of the error message is trimmed.)")
	}
	return strings.Join(lines, "\n")
}

func isStdout(w io.Writer) bool {
	if wc, ok := w.(*clc.NopWriteCloser); ok {
		return wc.W == os.Stdout
	}
	return false
}

func isStderr(w io.Writer) bool {
	if wc, ok := w.(*clc.NopWriteCloser); ok {
		return wc.W == os.Stderr
	}
	return false
}

// windowsCommandPrompt tries to determine if CLC is running in Command Prompt
func windowsCommandPrompt() bool {
	// WT_SESSION is a special environment variable set by Windows Terminal and it contains a GUID.
	wt := os.Getenv(envWtSession)
	isWt := len(wt) == 36 || wt == "clc"
	force := os.Getenv(envForceWindows) == "1"
	return (runtime.GOOS == "windows" || force) && !isWt
}

// fixStdoutStderr fixes stdout and stderr on Windows, so escape codes are not printed.
func fixStdoutStderr(stdout, stderr io.Writer) (io.Writer, io.Writer) {
	isCp := windowsCommandPrompt()
	if isStdout(stdout) {
		if isCp && color.NoColor {
			stdout = colorable.NewNonColorable(stdout)
		} else {
			// colorable.NewNonColorable doesn't work well on non-Windows
			stdout = colorable.NewColorableStdout()
		}
	}
	if isStderr(stderr) {
		if isCp && color.NoColor {
			stderr = colorable.NewNonColorable(stderr)
		} else {
			// colorable.NewNonColorable doesn't work well on non-Windows
			stderr = colorable.NewColorableStderr()
		}
	}
	return stdout, stderr
}
