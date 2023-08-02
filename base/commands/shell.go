//go:build std || shell

package commands

import (
	"context"
	"fmt"
	"strings"
	"sync"

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
	m, err := ec.(*cmd.ExecContext).Main().Clone(true)
	if err != nil {
		return fmt.Errorf("cloning Main: %w", err)
	}
	var cfgText, logText string
	if !terminal.IsPipe(ec.Stdin()) {
		cfgPathProp := ec.Props().GetString(clc.PropertyConfig)
		cfgPath = paths.ResolveConfigPath(cfgPathProp)
		if cfgPath != "" {
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
	endLineFn := makeEndLineFunc()
	textFn := makeTextFunc(m, ec, verbose, false, false, func(shortcut string) bool {
		cm.mu.RLock()
		_, ok := cm.shortcuts[shortcut]
		cm.mu.RUnlock()
		return ok
	})
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

func (*ShellCommand) Unwrappable() {}

func init() {
	Must(plug.Registry.RegisterCommand("shell", &ShellCommand{}))
}
