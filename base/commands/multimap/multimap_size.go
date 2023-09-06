//go:build std || multimap

package multimap

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

type MultiMapSizeCommand struct{}

func (mc *MultiMapSizeCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("size")
	help := "Return the size of the given MultiMap"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *MultiMapSizeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mmName := ec.Props().GetString(multiMapFlagName)
	mv, err := ec.Props().GetBlocking(multiMapPropertyName)
	if err != nil {
		return err
	}
	m := mv.(*hazelcast.MultiMap)
	sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Getting the size of the multimap %s", mmName))
		return m.Size(ctx)
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
	Must(plug.Registry.RegisterCommand("multi-map:size", &MultiMapSizeCommand{}))
}
