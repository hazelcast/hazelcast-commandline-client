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

const newVersionWarning = `
A newer version of CLC is available.

Visit the following link for release notes and to download:
https://github.com/hazelcast/hazelcast-commandline-client/releases/%s

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
		maybePrintNewerVersion(ec)
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

func maybePrintNewerVersion(ec plug.ExecContext) {
	if internal.IsCheckVersion == "disabled" {
		fmt.Println("I am not checking version")
		return
	}
	v, err := internal.LatestReleaseVersion()
	if err != nil {
		ec.Logger().Error(err)
		return
	}
	if isSkipNewerVersion() {
		return
	}
	if v != "" && internal.CheckVersion(trimVersion(v), ">", trimVersion(internal.Version)) {
		I2(fmt.Fprintf(ec.Stdout(), newVersionWarning, v))
	}
}

func isSkipNewerVersion() bool {
	return internal.Version == internal.UnknownVersion || strings.Contains(internal.Version, internal.CustomBuildSuffix)
}

func trimVersion(v string) string {
	return strings.TrimPrefix(strings.Split(v, "-")[0], "v")
}

func (*ShellCommand) Unwrappable() {}

func init() {
	Must(plug.Registry.RegisterCommand("shell", &ShellCommand{}))
}
