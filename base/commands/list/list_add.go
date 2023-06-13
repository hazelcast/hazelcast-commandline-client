//go:build base || list

package list

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type ListAddCommand struct{}

func (mc *ListAddCommand) Init(cc plug.InitContext) error {
	addValueTypeFlag(cc)
	cc.SetPositionalArgCount(1, 1)
	help := "Add a value in the given list"
	cc.AddIntFlag(listFlagIndex, "", -1, false, "index for the value")
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("add [value] [flags]")
	return nil
}

func (mc *ListAddCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	listName := ec.Props().GetString(listFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	// get the list just to ensure the corresponding proxy is created
	if _, err := ec.Props().GetBlocking(listPropertyName); err != nil {
		return err
	}
	valueStr := ec.Args()[0]
	vd, err := makeValueData(ec, ci, valueStr)
	if err != nil {
		return err
	}
	index := ec.Props().GetInt(listFlagIndex)
	var req *hazelcast.ClientMessage
	if index >= 0 {
		req = codec.EncodeListAddWithIndexRequest(listName, int32(index), vd)
	} else {
		req = codec.EncodeListAddRequest(listName, vd)
	}
	pid, err := stringToPartitionID(ci, listName)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Adding value into list %s", listName))
		return ci.InvokeOnPartition(ctx, req, pid, nil)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("list:add", &ListAddCommand{}))
}
