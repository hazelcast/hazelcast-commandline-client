package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type MapGetCommand struct{}

func (mc *MapGetCommand) Init(cc plug.InitContext) error {
	cc.AddStringFlag(mapFlagKeyType, "k", "", false, "key type")
	help := "Get a value from the given IMap"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("get -n MAP KEY [flags]")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (mc *MapGetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(mapFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	keyStr := ec.Args()[0]
	keyData, err := MakeKeyData(ec, ci, keyStr)
	if err != nil {
		return err
	}
	req := codec.EncodeMapGetRequest(mapName, keyData, 0)
	hint := fmt.Sprintf("Getting from map %s", mapName)
	rv, err := ec.ExecuteBlocking(ctx, hint, func(ctx context.Context) (any, error) {
		return ci.InvokeOnKey(ctx, req, keyData, nil)
	})
	if err != nil {
		return err
	}
	raw := codec.DecodeMapGetResponse(rv.(*hazelcast.ClientMessage))
	vt := raw.Type()
	value, err := ci.DecodeData(raw)
	if err != nil {
		ec.Logger().Info("The value for %s was not decoded, due to error: %s", keyStr, err.Error())
		value = serialization.NondecodedType(serialization.TypeToString(vt))
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
			Value: serialization.TypeToString(vt),
		})
	}
	ec.AddOutputRows(row)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:get", &MapGetCommand{}))
}
