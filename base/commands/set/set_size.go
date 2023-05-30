package _set

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-go-client"
)

type SetSizeCommand struct{}

func (sc *SetSizeCommand) Init(cc plug.InitContext) error {
	help := "Return the size of the given Set"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("size [flags]")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (sc *SetSizeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	setName := ec.Props().GetString(setFlagName)
	qv, err := ec.Props().GetBlocking(setPropertyName)
	if err != nil {
		return err
	}
	s := qv.(*hazelcast.Set)
	sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Getting the size of the set %s", setName))
		return s.Size(ctx)
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, output.Row{
		{
			Name:  "Size",
			Type:  serialization.TypeInt32,
			Value: int32(sv.(int)),
		},
	})
}

func init() {
	Must(plug.Registry.RegisterCommand("set:size", &SetSizeCommand{}))
}
