//go:build base || multimap

package _multimap

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-go-client"
)

type MultiMapValuesCommand struct{}

func (m MultiMapValuesCommand) Init(cc plug.InitContext) error {
	help := "Get all values of a MultiMap"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("values")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (m MultiMapValuesCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mmName := ec.Props().GetString(multiMapFlagName)
	showType := ec.Props().GetBool(multiMapFlagShowType)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	req := codec.EncodeMultiMapValuesRequest(mmName)
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Getting values of %s", mmName))
		return ci.InvokeOnRandomTarget(ctx, req, nil)
	})
	if err != nil {
		return err
	}
	stop()
	raw := codec.DecodeMultiMapValuesResponse(rv.(*hazelcast.ClientMessage))
	var rows []output.Row
	for _, r := range raw {
		t := r.Type()
		v, err := ci.DecodeData(*r)
		if err != nil {
			v = serialization.NondecodedType(serialization.TypeToLabel(t))
		}
		row := output.Row{
			output.Column{
				Name:  output.NameValue,
				Type:  t,
				Value: v,
			},
		}
		if showType {
			row = append(row, output.Column{
				Name:  output.NameValueType,
				Type:  serialization.TypeString,
				Value: serialization.TypeToLabel(t),
			})
		}
		rows = append(rows, row)
	}
	if len(rows) > 0 {
		return ec.AddOutputRows(ctx, rows...)
	}
	ec.PrintlnUnnecessary("No values found.")
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("multi-map:values", &MultiMapValuesCommand{}))
}
