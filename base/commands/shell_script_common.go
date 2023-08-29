//go:build std || script || shell

package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/shlex"

	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	flagIgnoreErrors = "ignore-errors"
	flagEcho         = "echo"
)

type shortcutFunc func(shortcut string) bool

func makeEndLineFunc() shell.EndLineFn {
	clcMultilineContinue := false
	return func(line string, multiline bool) (string, bool) {
		// not caching trimmed line, since we want the backslash at the very end of the line. --YT
		clcCmd := strings.HasPrefix(strings.TrimSpace(line), shell.CmdPrefix)
		if clcCmd || multiline && clcMultilineContinue {
			clcMultilineContinue = true
			end := !strings.HasSuffix(line, "\\")
			if !end {
				line = line[:len(line)-1]
			}
			return line, end
		}
		clcMultilineContinue = false
		lt := strings.TrimSpace(line)
		end := strings.HasPrefix(lt, "help") || strings.HasPrefix(lt, shell.CmdPrefix) || strings.HasSuffix(lt, ";")
		if !end {
			line += "\n"
		}
		return line, end
	}
}

func makeTextFunc(m *cmd.Main, ec plug.ExecContext, verbose, ignoreErrors, echo bool, sf shortcutFunc) shell.TextFn {
	return func(ctx context.Context, stdout io.Writer, text string) error {
		if strings.HasPrefix(strings.TrimSpace(text), shell.CmdPrefix) {
			parts := strings.Fields(text)
			ok := sf(parts[0])
			if !ok {
				// this is a CLC command
				text = strings.TrimSpace(text)
				text = strings.TrimPrefix(text, shell.CmdPrefix)
				args, err := shlex.Split(text)
				if err != nil {
					return err
				}
				args[0] = shell.CmdPrefix + args[0]
				return m.Execute(ctx, args...)
			}
		}
		f, err := shell.ConvertStatement(ctx, ec, text, verbose)
		if err != nil {
			if errors.Is(err, shell.ErrHelp) {
				check.I2(fmt.Fprintln(stdout, shell.InteractiveHelp()))
				return nil
			}
			return err
		}
		if w, ok := ec.(plug.ResultWrapper); ok {
			return w.WrapResult(f)
		}
		return f()
	}
}
