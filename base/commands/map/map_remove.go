//go:build base || map

package _map

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

type MapRemoveCommand struct{}

func (mc *MapRemoveCommand) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	help := "Remove a value from the given Map"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("remove [-n MAP] KEY [flags]")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (mc *MapRemoveCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(mapFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	keyStr := ec.Args()[0]
	keyData, err := makeKeyData(ec, ci, keyStr)
	if err != nil {
		return err
	}
	req := codec.EncodeMapRemoveRequest(mapName, keyData, 0)
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Removing from map %s", mapName))
		return ci.InvokeOnKey(ctx, req, keyData, nil)
	})
	if err != nil {
		return err
	}
	stop()
	raw := codec.DecodeMapRemoveResponse(rv.(*hazelcast.ClientMessage))
	vt := raw.Type()
	value, err := ci.DecodeData(raw)
	if err != nil {
		ec.Logger().Info("The value for %s was not decoded, due to error: %s", keyStr, err.Error())
		value = serialization.NondecodedType(serialization.TypeToLabel(vt))
	}
	row := output.Row{
		output.Column{
			Name:  output.NameValue,
			Type:  vt,
			Value: value,
		},
	}
	if ec.Props().GetBool(mapFlagShowType) {
		row = append(row, output.Column{
			Name:  output.NameValueType,
			Type:  serialization.TypeString,
			Value: serialization.TypeToLabel(vt),
		})
	}
	return ec.AddOutputRows(ctx, row)
}

func init() {
	Must(plug.Registry.RegisterCommand("map:remove", &MapRemoveCommand{}))
}
