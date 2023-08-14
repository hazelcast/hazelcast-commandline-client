package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type OneshotShell struct {
	endLineFn     EndLineFn
	textFn        TextFn
	commentPrefix string
	stderr        io.Writer
	stdout        io.Writer
	stdin         io.Reader
	ignoreErrors  bool
	echo          bool
}

func NewOneshotShell(endLineFn EndLineFn, sio clc.IO, textFn TextFn) *OneshotShell {
	return &OneshotShell{
		endLineFn:     endLineFn,
		textFn:        textFn,
		commentPrefix: "",
		stderr:        sio.Stderr,
		stdout:        sio.Stdout,
		stdin:         sio.Stdin,
	}
}

func (sh *OneshotShell) SetCommentPrefix(pfx string) {
	sh.commentPrefix = pfx
}

func (sh *OneshotShell) SetIgnoreErrors(ignore bool) {
	sh.ignoreErrors = ignore
}

func (sh *OneshotShell) SetEcho(echo bool) {
	sh.echo = echo
}

func (sh *OneshotShell) Run(ctx context.Context) error {
	if err := sh.readTextBasic(ctx); err != nil {
		return err
	}
	return nil
}

func (sh *OneshotShell) readTextBasic(ctx context.Context) error {
	// NOTE: when this implementation is changed,
	// clc/shell/shell.go:readTextReadline should also change!
	var sb strings.Builder
	multiline := false
	var line int
	var multilineStart int
	var prevMultiline bool
	echo := func(text string) {
		if sh.echo {
			at := line
			if prevMultiline {
				at = multilineStart
			}
			check.I2(fmt.Fprint(sh.stderr, fmt.Sprintf("%d: %s\n", at, text)))
		}
	}
	scn := bufio.NewScanner(sh.stdin)
	for scn.Scan() {
		if scn.Err() != nil {
			return scn.Err()
		}
		line++
		p := scn.Text()
		if !multiline {
			pt := strings.TrimSpace(p)
			if pt == "" {
				continue
			}
			if sh.commentPrefix != "" && strings.HasPrefix(pt, sh.commentPrefix) {
				continue
			}
		}
		text, end := sh.endLineFn(p, multiline)
		sb.WriteString(text)
		prevMultiline = multiline
		multiline = !end
		if !prevMultiline && multiline {
			// multiline commands should have the first line displayed in echo
			multilineStart = line
		}
		if end {
			echo(sb.String())
			if err := sh.textFn(ctx, sh.stdout, sb.String()); err != nil {
				if !sh.ignoreErrors {
					return err
				}
			}
			sb.Reset()
		}
	}
	if sb.String() != "" {
		echo(sb.String())
		err := sh.textFn(ctx, sh.stdout, sb.String())
		if err != nil {
			if !sh.ignoreErrors {
				return err
			}
		}
	}
	return nil
}
