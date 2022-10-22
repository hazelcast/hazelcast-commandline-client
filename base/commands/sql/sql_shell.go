package sql

import (
	"context"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type SQLShellCommand struct {
	client  *hazelcast.Client
	verbose bool
}

func (cm *SQLShellCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	if ec.CommandName() == "clc sql shell" {
		props.Set(clc.PropertyOutputFormat, "table")
	}
	return nil
}

func (cm *SQLShellCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("shell")
	return nil
}

func (cm *SQLShellCommand) Exec(ec plug.ExecContext) error {
	ctx := context.Background()
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	cm.client = ci.Client()
	cm.verbose = ec.Props().GetBool(clc.PropertyVerbose)
	return nil
}

func (cm *SQLShellCommand) ExecInteractive(ec plug.ExecInteractiveContext) error {
	sh := shell.NewShell("SQL> ", "... ", "",
		ec.Stdout(), ec.Stderr(),
		func(line string) bool {
			return strings.HasSuffix(line, ";")
		},
		func(ctx context.Context, text string) error {
			res, err := cm.client.SQL().Execute(ctx, text)
			if err != nil {
				return adaptSQLError(err)
			}
			if err := updateOutput(ec, res, cm.verbose); err != nil {
				return err
			}
			if err := ec.FlushOutput(); err != nil {
				return err
			}
			return nil
		},
	)
	defer sh.Close()
	return sh.Start(context.Background())
}

func init() {
	plug.Registry.RegisterAugmentor("20-sql-shell", &SQLShellCommand{})
	Must(plug.Registry.RegisterCommand("sql:shell", &SQLShellCommand{}))
}
