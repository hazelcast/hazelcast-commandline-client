//go:build std || list

package list

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type ListSetCommand struct{}

func (mc *ListSetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("set")
	help := "Set a value at the given index in the list"
	cc.SetCommandHelp(help, help)
	addValueTypeFlag(cc)
	cc.AddInt64Arg(argIndex, argTitleIndex)
	cc.AddStringArg(argValue, argTitleValue)
	return nil
}

func (mc *ListSetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(listFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	// get the list just to ensure the corresponding proxy is created
	if _, err := ec.Props().GetBlocking(listPropertyName); err != nil {
		return err
	}
	index := ec.GetInt64Arg(argIndex)
	valueStr := ec.GetStringArg(argValue)
	vd, err := makeValueData(ec, ci, valueStr)
	if err != nil {
		return err
	}
	pid, err := stringToPartitionID(ci, name)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Setting the value of the list %s", name))
		req := codec.EncodeListSetRequest(name, int32(index), vd)
		return ci.InvokeOnPartition(ctx, req, pid, nil)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("list:set", &ListSetCommand{}))
}
