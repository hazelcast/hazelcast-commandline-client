//go:build std || list

package list

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type AddCommand struct{}

func (AddCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("add")
	help := "Add a value in the given list"
	cc.SetCommandHelp(help, help)
	commands.AddValueTypeFlag(cc)
	cc.AddIntFlag(flagIndex, "", -1, false, "index for the value")
	cc.AddStringArg(base.ArgValue, base.ArgTitleValue)
	return nil
}

func (AddCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	val, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		// get the list just to ensure the corresponding proxy is created
		_, err = getList(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		valueStr := ec.GetStringArg(base.ArgValue)
		vd, err := commands.MakeValueData(ec, ci, valueStr)
		if err != nil {
			return nil, err
		}
		index := ec.Props().GetInt(flagIndex)
		var req *hazelcast.ClientMessage
		if index >= 0 {
			req = codec.EncodeListAddWithIndexRequest(name, int32(index), vd)
		} else {
			req = codec.EncodeListAddRequest(name, vd)
		}
		pid, err := internal.StringToPartitionID(ci, name)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Adding value at index %d into List '%s'", index, name))
		resp, err := ci.InvokeOnPartition(ctx, req, pid, nil)
		if err != nil {
			return nil, err
		}
		if index >= 0 {
			return true, nil
		}
		return codec.DecodeListAddResponse(resp), nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Updated List '%s'.\n", name)
	ec.PrintlnUnnecessary(msg)
	row := output.Row{
		output.Column{
			Name:  "Value Changed",
			Type:  serialization.TypeBool,
			Value: val,
		},
	}
	return ec.AddOutputRows(ctx, row)
}

func init() {
	Must(plug.Registry.RegisterCommand("list:add", &AddCommand{}))
}
