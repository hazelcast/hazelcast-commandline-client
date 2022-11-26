//go:build base

package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/shlex"

	"github.com/hazelcast/hazelcast-commandline-client/base/commands/sql"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	puberrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

var errHelp = errors.New("interactive help")

type ShellCommand struct {
	shortcuts map[string]struct{}
}

func (cm *ShellCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("shell")
	help := "Start the interactive shell"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, 0)
	cm.shortcuts = map[string]struct{}{
		`\dm`:   {},
		`\dm+`:  {},
		`\exit`: {},
	}
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
	var cfgText string
	cfgPath := ec.Props().GetString(clc.PropertyConfig)
	if cfgPath != "" {
		cfgText = fmt.Sprintf("Using configuration at: %s\n", cfgPath)
	}
	if !shell.IsPipe() {
		I2(fmt.Fprintf(ec.Stdout(), `Hazelcast CLC %s (c) 2022 Hazelcast Inc.
		
Participate to our survey at: https://forms.gle/rPFywdQjvib1QCe49

%sType 'help' for help information. Prefix non-SQL commands with \
	
	`, internal.Version, cfgText))
	}
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	clcMultilineContinue := false
	endLineFn := func(line string, multiline bool) (string, bool) {
		// not caching trimmed line, since we want the backslash at the very end of the line. --YT
		clcCmd := strings.HasPrefix(strings.TrimSpace(line), "\\")
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
		end := strings.HasPrefix(line, "help") || strings.HasPrefix(line, "\\") || strings.HasSuffix(line, ";")
		if !end {
			line = fmt.Sprintf("%s\n", line)
		}
		return line, end
	}
	textFn := func(ctx context.Context, text string) error {
		if strings.HasPrefix(strings.TrimSpace(text), "\\") {
			parts := strings.Fields(text)
			if _, ok := cm.shortcuts[parts[0]]; !ok {
				// this is a CLC command
				text = strings.TrimSpace(text)
				text = strings.TrimPrefix(text, "\\")
				args, err := shlex.Split(text)
				if err != nil {
					return err
				}
				args[0] = fmt.Sprintf("\\%s", args[0])
				return m.Execute(args)
			}
		}
		text, err := convertStatement(text)
		if err != nil {
			if errors.Is(err, errHelp) {
				I2(fmt.Fprintln(ec.Stdout(), interactiveHelp()))
				return nil
			}
			return err
		}
		res, stop, err := sql.ExecSQL(ctx, ec, text)
		if err != nil {
			return err
		}
		defer stop()
		if err := sql.UpdateOutput(ctx, ec, res, verbose); err != nil {
			return err
		}
		return nil
	}
	path := paths.Join(paths.Home(), "shell.history")
	if shell.IsPipe() {
		sh := shell.NewOneshot(endLineFn, textFn)
		sh.SetCommentPrefix("--")
		return sh.Run(context.Background())
	}
	sh, err := shell.New("CLC> ", " ... ", path, "sql", ec.Stdout(), ec.Stderr(), endLineFn, textFn)
	if err != nil {
		return err
	}
	sh.SetCommentPrefix("--")
	defer sh.Close()
	return sh.Start(ctx)
}

func convertStatement(stmt string) (string, error) {
	stmt = strings.TrimSpace(stmt)
	if strings.HasPrefix(stmt, "help") {
		return "", errHelp
	}
	if strings.HasPrefix(stmt, "\\") {
		// this is a shell command
		parts := strings.Fields(stmt)
		switch parts[0] {
		case "\\dm":
			if len(parts) == 1 {
				return "show mappings;", nil
			}
			if len(parts) == 2 {
				// escape single quote
				mn := strings.Replace(parts[1], "'", "''", -1)
				return fmt.Sprintf(`
					SELECT * FROM information_schema.mappings
					WHERE table_name = '%s';
				`, mn), nil
			}
			return "", fmt.Errorf("Usage: \\dm [mapping]")
		case "\\dm+":
			if len(parts) == 1 {
				return "show mappings;", nil
			}
			if len(parts) == 2 {
				// escape single quote
				mn := strings.Replace(parts[1], "'", "''", -1)
				return fmt.Sprintf(`
					SELECT * FROM information_schema.columns
					WHERE table_name = '%s';
				`, mn), nil
			}
			return "", fmt.Errorf("Usage: \\dm+ [mapping]")
		case "\\exit":
			return "", shell.ErrExit
		}
		return "", fmt.Errorf("Unknown shell command: %s", stmt)
	}
	return stmt, nil
}

func interactiveHelp() string {
	return `
Shortcut Commands:
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
