//go:build std || atomiclong

package atomiclong

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type SetCommand struct{}

func (mc *SetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("set")
	help := "Set the value of the AtomicLong"
	cc.SetCommandHelp(help, help)
	cc.AddInt64Arg(base.ArgValue, base.ArgTitleValue)
	return nil
}

func (mc *SetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	row, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) (output.Row, error) {
		ali, err := getAtomicLong(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Setting value of AtomicLong %s", name))
		v := ec.GetInt64Arg(base.ArgValue)
		err = ali.Set(ctx, v)
		if err != nil {
			return nil, err
		}
		s := executeState{
			Name:  name,
			Value: v,
		}
		row := output.Row{
			output.Column{
				Name:  "Value",
				Type:  serialization.TypeInt64,
				Value: s.Value,
			},
		}
		return row, nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Set AtomicLong %s.\n", name)
	ec.PrintlnUnnecessary(msg)
	return ec.AddOutputRows(ctx, row)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("atomic-long:set", &SetCommand{}))
}
