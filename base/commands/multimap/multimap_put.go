//go:build std || multimap

package multimap

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type MultiMapPutCommand struct{}

func (MultiMapPutCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("put")
	help := "Put a value in the given MultiMap"
	cc.SetCommandHelp(help, help)
	commands.AddKeyTypeFlag(cc)
	commands.AddValueTypeFlag(cc)
	cc.AddStringArg(commands.ArgKey, commands.ArgTitleKey)
	cc.AddStringArg(base.ArgValue, base.ArgTitleValue)
	return nil
}

func (MultiMapPutCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	keyStr := ec.GetStringArg(commands.ArgKey)
	valueStr := ec.GetStringArg(base.ArgValue)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		cid, vid := cmd.FindClusterIDs(ctx, ec)
		ec.Metrics().Increment(metrics.NewKey(cid, vid), "total.multimap."+cmd.RunningModeString(ec))
		sp.SetText(fmt.Sprintf("Putting value into MultiMap '%s'", name))
		kd, vd, err := commands.MakeKeyValueData(ec, ci, keyStr, valueStr)
		if err != nil {
			return nil, err
		}
		req := codec.EncodeMultiMapPutRequest(name, kd, vd, 0)
		resp, err := ci.InvokeOnKey(ctx, req, kd, nil)
		if err != nil {
			return nil, err
		}
		value := codec.DecodeMultiMapPutResponse(resp)
		row := output.Row{
			output.Column{
				Name:  output.NameValue,
				Type:  serialization.TypeBool,
				Value: value,
			},
		}
		if ec.Props().GetBool(base.FlagShowType) {
			row = append(row, output.Column{
				Name:  output.NameValueType,
				Type:  serialization.TypeString,
				Value: serialization.TypeToLabel(serialization.TypeBool),
			})
		}
		return []output.Row{row}, nil
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, rowsV.([]output.Row)...)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("multi-map:put", &MultiMapPutCommand{}))
}
