//go:build base || list

package list

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListSizeCommand struct{}

func (mc *ListSizeCommand) Init(cc plug.InitContext) error {
	help := "Return the size of the given List"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("size [flags]")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (mc *ListSizeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	listName := ec.Props().GetString(listFlagName)
	lv, err := ec.Props().GetBlocking(listPropertyName)
	if err != nil {
		return err
	}
	l := lv.(*hazelcast.List)
	sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Getting the size of list %s", listName))
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
