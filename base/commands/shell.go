//go:build base

package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/google/shlex"
	_map "github.com/hazelcast/hazelcast-commandline-client/base/commands/map"

	"github.com/hazelcast/hazelcast-commandline-client/base/commands/sql"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	puberrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/terminal"
)

const banner = `Hazelcast CLC %s (c) 2023 Hazelcast Inc.
		
* Participate in our survey at: https://forms.gle/rPFywdQjvib1QCe49
* Type 'help' for help information. Prefix non-SQL commands with \
		
%s%s

`

var errHelp = errors.New("interactive help")

type ShellCommand struct {
	shortcuts map[string]struct{}
	mu        sync.RWMutex
}

func (cm *ShellCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("shell")
	help := "Start the interactive shell"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, 0)
	cc.Hide()
	cm.mu.Lock()
	cm.shortcuts = map[string]struct{}{
		`\di`:   {},
		`\dm`:   {},
		`\dm+`:  {},
		`\exit`: {},
	}
	cm.mu.Unlock()
	return nil
}

func (cm *ShellCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (cm *ShellCommand) ExecInteractive(ctx context.Context, ec plug.ExecContext) error {
	if len(ec.Args()) > 0 {
		return puberrors.ErrNotAvailable
	}
	m, err := ec.(*cmd.ExecContext).Main().CloneForInteractiveMode()
	if err != nil {
		return fmt.Errorf("cloning Main: %w", err)
	}
	var cfgText, logText string
	if !terminal.IsPipe(ec.Stdin()) {
		cfgPath := ec.Props().GetString(clc.PropertyConfig)
		if cfgPath != "" {
			cfgPath = paths.ResolveConfigPath(cfgPath)
			cfgText = fmt.Sprintf("Configuration : %s\n", cfgPath)
		}
		logPath := ec.Props().GetString(clc.PropertyLogPath)
		if logPath != "" {
			logLevel := strings.ToUpper(ec.Props().GetString(clc.PropertyLogLevel))
			logText = fmt.Sprintf("Log %9s : %s", logLevel, logPath)
		}
		I2(fmt.Fprintf(ec.Stdout(), banner, internal.Version, cfgText, logText))
	}
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	clcMultilineContinue := false
	endLineFn := func(line string, multiline bool) (string, bool) {
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
		line = strings.TrimSpace(line)
		end := strings.HasPrefix(line, "help") || strings.HasPrefix(line, shell.CmdPrefix) || strings.HasSuffix(line, ";")
		if !end {
			line = fmt.Sprintf("%s\n", line)
		}
		return line, end
	}
	textFn := func(ctx context.Context, stdout io.Writer, text string) error {
		if strings.HasPrefix(strings.TrimSpace(text), shell.CmdPrefix) {
			parts := strings.Fields(text)
			cm.mu.RLock()
			_, ok := cm.shortcuts[parts[0]]
			cm.mu.RUnlock()
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
		f, err := convertStatement(ctx, ec, text, verbose)
		if err != nil {
			if errors.Is(err, errHelp) {
				I2(fmt.Fprintln(stdout, interactiveHelp()))
				return nil
			}
			return err
		}
		if w, ok := ec.(plug.ResultWrapper); ok {
			return w.WrapResult(f)
		}
		return f()
	}
	path := paths.Join(paths.Home(), "shell.history")
	if terminal.IsPipe(ec.Stdin()) {
		sio := clc.IO{
			Stdin:  ec.Stdin(),
			Stderr: ec.Stderr(),
			Stdout: ec.Stdout(),
		}
		sh := shell.NewOneshotShell(endLineFn, sio, textFn)
		sh.SetCommentPrefix("--")
		return sh.Run(ctx)
	}
	sh, err := shell.New("CLC> ", " ... ", path, ec.Stdout(), ec.Stderr(), ec.Stdin(), endLineFn, textFn)
	if err != nil {
		return err
	}
	sh.SetCommentPrefix("--")
	defer sh.Close()
	return sh.Start(ctx)
}

func (ShellCommand) Unwrappable() {}

func convertStatement(ctx context.Context, ec plug.ExecContext, stmt string, verbose bool) (func() error, error) {
	var query string
	stmt = strings.TrimSpace(stmt)
	if strings.HasPrefix(stmt, "help") {
		return nil, errHelp
	}
	if strings.HasPrefix(stmt, shell.CmdPrefix) {
		// this is a shell command
		stmt = strings.TrimPrefix(stmt, "\\")
		parts := strings.Fields(stmt)
		switch parts[0] {
		case "di":
			return func() error {
				return _map.Exec(ctx, ec)
			}, nil
		case "dm":
			if len(parts) == 1 {
				query = "show mappings;"
				return nil, nil
			} else if len(parts) == 2 {
				// escape single quote
				mn := strings.Replace(parts[1], "'", "''", -1)
				query = fmt.Sprintf(`
					SELECT * FROM information_schema.mappings
					WHERE table_name = '%s';
				`, mn)
			} else {
				return nil, fmt.Errorf("Usage: %sdm [mapping]", shell.CmdPrefix)
			}
		case "dm+":
			if len(parts) == 1 {
				query = "show mappings;"
				return nil, nil
			} else if len(parts) == 2 {
				// escape single quote
				mn := strings.Replace(parts[1], "'", "''", -1)
				query = fmt.Sprintf(`
					SELECT * FROM information_schema.columns
					WHERE table_name = '%s';
				`, mn)
				return nil, nil
			} else {
				return nil, fmt.Errorf("Usage: %sdm+ [mapping]", shell.CmdPrefix)
			}
		case "exit":
			return nil, shell.ErrExit
		default:
			return nil, fmt.Errorf("Unknown shell command: %s", stmt)
		}
	}
	f := func() error {
		res, stop, err := sql.ExecSQL(ctx, ec, query)
		if err != nil {
			return err
		}
		defer stop()
		// TODO: update sql.UpdateOutput to use stdout
		if err := sql.UpdateOutput(ctx, ec, res, verbose); err != nil {
			return err
		}
		return nil
	}
	return f, nil
}

func interactiveHelp() string {
	return `
Shortcut Commands:
	\di			  List Indexes
	\di	 MAPPING  List Indexes for a specific map
	\dm           List mappings
	\dm  MAPPING  Display information about a mapping
	\dm+ MAPPING  Describe a mapping
	\exit         Exit the shell
	\help         Display help for CLC commands
`
}

func init() {
	Must(plug.Registry.RegisterCommand("shell", &ShellCommand{}))
}
