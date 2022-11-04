package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/shlex"

	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ShellCommand struct{}

func (cm *ShellCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("shell")
	help := "Start the interactive shell"
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm *ShellCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (cm *ShellCommand) ExecInteractive(ctx context.Context, ec plug.ExecContext) error {
	m, err := ec.(*cmd.ExecContext).Main().CloneForInteractiveMode()
	if err != nil {
		return fmt.Errorf("cloning Main: %w", err)
	}
	endLineFn := func(line string) (string, bool) {
		end := !strings.HasSuffix(line, "\\")
		if !end {
			line = line[:len(line)-1]
		}
		return line, end
	}
	textFn := func(ctx context.Context, text string) error {
		text = strings.TrimSpace(text)
		args, err := shlex.Split(text)
		if err != nil {
			return err
		}
		return m.Execute(args)
	}
	path := paths.Join(paths.Home(), "shell.history")
	if shell.IsPipe() {
		// set interactive mode to false, so the animations and other stuff doesn't affect the input
		ec.(*cmd.ExecContext).SetInteractive(false)
		sh := shell.NewOneshot(endLineFn, textFn)
		sh.SetCommentPrefix("#")
		return sh.Run(context.Background())
	}
	sh := shell.New("CLC> ", " ... ", path, ec.Stdout(), ec.Stderr(), endLineFn, textFn)
	sh.SetCommentPrefix("#")
	defer sh.Close()
	return sh.Start(context.Background())
}

func init() {
	Must(plug.Registry.RegisterCommand("shell", &ShellCommand{}))
}
