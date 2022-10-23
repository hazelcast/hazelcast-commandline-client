package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/shlex"
	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ShellCommand struct {
	ci *hazelcast.ClientInternal
}

func (cm *ShellCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("shell")
	help := "Start the interactive shell"
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm *ShellCommand) Exec(ec plug.ExecContext) error {
	ctx := context.Background()
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	cm.ci = ci
	return nil
}

func (cm *ShellCommand) ExecInteractive(ec plug.ExecInteractiveContext) error {
	m, err := ec.(*cmd.ExecContext).Main().CloneForInteractiveMode()
	if err != nil {
		return fmt.Errorf("cloning Main: %w", err)
	}
	endLineFn := func(line string) bool {
		return !strings.HasSuffix(line, "\\")
	}
	textFn := func(ctx context.Context, text string) error {
		args, err := shlex.Split(text)
		if err != nil {
			return err
		}
		return m.Execute(args)
	}
	sh := shell.New("CLC> ", "... ", "",
		ec.Stdout(), ec.Stderr(), endLineFn, textFn)
	defer sh.Close()
	return sh.Start(context.Background())
}

func init() {
	Must(plug.Registry.RegisterCommand("shell", &ShellCommand{}))
}
