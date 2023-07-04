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
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListContainsCommand struct{}

func (mc *ListContainsCommand) Init(cc plug.InitContext) error {
	addValueTypeFlag(cc)
	cc.SetPositionalArgCount(1, 1)
	help := "Check if the value is present in the list."
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("contains [value] [flags]")
	return nil
}

func (mc *ListContainsCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(listFlagName)
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
	pid, err := stringToPartitionID(ci, name)
	if err != nil {
		return err
	}
	cmi, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Checking if value exists in the list %s", name))
		req := codec.EncodeListContainsRequest(name, vd)
		return ci.InvokeOnPartition(ctx, req, pid, nil)
	})
	if err != nil {
		return err
	}
	stop()
	cm := cmi.(*hazelcast.ClientMessage)
	contains := codec.DecodeListContainsResponse(cm)
	return ec.AddOutputRows(ctx, output.Row{
		{
			Name:  "Contains",
			Type:  serialization.TypeBool,
			Value: contains,
		},
	})
}

func init() {
	Must(plug.Registry.RegisterCommand("list:contains", &ListContainsCommand{}))
}
