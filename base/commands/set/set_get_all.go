package _set

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-go-client"
)

type SetGetAllCommand struct{}

func (sc *SetGetAllCommand) Init(cc plug.InitContext) error {
	help := "Return the elements of the given Set"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("get-all [flags]")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (sc *SetGetAllCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	setName := ec.Props().GetString(setFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	req := codec.EncodeSetGetAllRequest(setName)
	pID, err := stringToPartitionID(ci, setName)
	if err != nil {
		return err
	}
	sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Removing from set %s", setName))
		return ci.InvokeOnPartition(ctx, req, pID, nil)
	})
	if err != nil {
		return err
	}
	stop()
	raw := codec.DecodeSetGetAllResponse(sv.(*hazelcast.ClientMessage))
	var rows []output.Row
	for _, r := range raw {
		val, err := ci.DecodeData(*r)
		if err != nil {
			return err
		}
		rows = append(rows, output.Row{
			{
				Name:  "element",
				Type:  r.Type(),
				Value: val,
			},
		})
	}
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	Must(plug.Registry.RegisterCommand("set:get-all", &SetGetAllCommand{}))
}
