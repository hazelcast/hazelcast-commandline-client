//go:build std || list

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
	cc.SetCommandUsage("add")
	help := "Add a value in the given list"
	cc.SetCommandHelp(help, help)
	addValueTypeFlag(cc)
	cc.AddIntFlag(listFlagIndex, "", -1, false, "index for the value")
	cc.AddStringArg(argValue, argTitleValue)
	return nil
}

func (mc *ListAddCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(listFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	// get the list just to ensure the corresponding proxy is created
	if _, err := ec.Props().GetBlocking(listPropertyName); err != nil {
		return err
	}
	valueStr := ec.GetStringArg(argValue)
	vd, err := makeValueData(ec, ci, valueStr)
	if err != nil {
		return err
	}
	index := ec.Props().GetInt(listFlagIndex)
	var req *hazelcast.ClientMessage
	if index >= 0 {
		req = codec.EncodeListAddWithIndexRequest(name, int32(index), vd)
	} else {
		req = codec.EncodeListAddRequest(name, vd)
	}
	pid, err := stringToPartitionID(ci, name)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Adding value at index %d into list %s", index, name))
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
