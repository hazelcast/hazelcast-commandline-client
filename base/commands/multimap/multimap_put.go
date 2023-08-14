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
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type MultiMapPutCommand struct {
}

func (m MultiMapPutCommand) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	addValueTypeFlag(cc)
	cc.SetPositionalArgCount(2, 2)
	help := "Put a value in the given MultiMap"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("put [key] [value] [flags]")
	return nil
}

func (m MultiMapPutCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mmName := ec.Props().GetString(multiMapFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	if _, err := ec.Props().GetBlocking(multiMapPropertyName); err != nil {
		return err
	}
	keyStr := ec.Args()[0]
	valueStr := ec.Args()[1]
	kd, vd, err := makeKeyValueData(ec, ci, keyStr, valueStr)
	if err != nil {
		return err
	}
	req := codec.EncodeMultiMapPutRequest(mmName, kd, vd, 0)
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Putting value into multimap %s", mmName))
		return ci.InvokeOnKey(ctx, req, kd, nil)
	})
	if err != nil {
		return err
	}
	stop()
	resp := codec.DecodeMultiMapPutResponse(rv.(*hazelcast.ClientMessage))
	row := output.Row{
		output.Column{
			Name:  output.NameValue,
			Type:  serialization.TypeBool,
			Value: resp,
		},
	}
	if ec.Props().GetBool(multiMapFlagShowType) {
		row = append(row, output.Column{
			Name:  output.NameValueType,
			Type:  serialization.TypeString,
			Value: serialization.TypeToLabel(serialization.TypeBool),
		})
	}
	return ec.AddOutputRows(ctx, row)
}

func init() {
	Must(plug.Registry.RegisterCommand("multi-map:put", &MultiMapPutCommand{}))
}
