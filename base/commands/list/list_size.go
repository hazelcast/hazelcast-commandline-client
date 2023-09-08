//go:build std || list

package list

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListSizeCommand struct{}

func (mc *ListSizeCommand) Unwrappable() {}

func (mc *ListSizeCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("size")
	help := "Return the size of the given List"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *ListSizeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		l, err := getList(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Getting the size of list %s", name))
		return l.Size(ctx)
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
	Must(plug.Registry.RegisterCommand("list:size", &ListSizeCommand{}))
}
