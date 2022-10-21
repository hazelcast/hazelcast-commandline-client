package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/peterh/liner"

	"github.com/hazelcast/hazelcast-commandline-client/clc/property"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type SQLShellCommand struct {
	client  *hazelcast.Client
	verbose bool
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
	cm.verbose = ec.Props().GetBool(property.Verbose)
	return nil
}

func (cm *SQLShellCommand) ExecInteractive(ec plug.ExecInteractiveContext) error {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	line.SetMultiLineMode(true)
	ctx := context.Background()
	for {
		ok, err := cm.execSQL(ctx, ec, line)
		if !ok {
			return nil
		}
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func (cm *SQLShellCommand) execSQL(ctx context.Context, ec plug.ExecInteractiveContext, line *liner.State) (bool, error) {
	prompt := "> "
	var sb strings.Builder
	for {
		p, err := line.Prompt(prompt)
		if err == liner.ErrPromptAborted {
			return false, err
		}
		if err != nil {
			return false, err
		}
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		sb.WriteString(p)
		sb.WriteString("\n")
		if strings.HasSuffix(p, ";") {
			break
		}
		prompt = "... "
	}
	query := sb.String()
	line.AppendHistory(query)
	res, err := cm.client.SQL().Execute(ctx, query)
	if err != nil {
		return true, adaptSQLError(err)
	}
	if err := updateOutput(ec, res, cm.verbose); err != nil {
		return true, err
	}
	return true, ec.FlushOutput()
}

func init() {
	Must(plug.Registry.RegisterCommand("sql:shell", &SQLShellCommand{}))
}
