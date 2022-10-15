package commands

import (
	"context"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const (
	mapFlagKey = "key"
)

type MapGetCommand struct{}

func (mc *MapGetCommand) Init(ctx plug.CommandContext) error {
	ctx.AddStringFlag(mapFlagKey, "k", "", true, "IMap key")
	return nil
}

func (mc *MapGetCommand) Exec(ec plug.ExecContext) error {
	ctx := context.TODO()
	key := ec.Props().GetString(mapFlagKey)
	mapName := ec.Props().GetString(mapFlagName)
	ci := MustAnyValue[*hazelcast.ClientInternal](ec.Props().GetBlocking(base.PropertyClientInternalName))
	keyData, err := ci.EncodeData(key)
	if err != nil {
		return err
	}
	req := codec.EncodeMapGetRequest(mapName, keyData, 0)
	resp, err := ci.InvokeOnKey(ctx, req, keyData, nil)
	if err != nil {
		return err
	}
	raw := codec.DecodeMapGetResponse(resp)
	valueType := raw.Type()
	value, err := ci.DecodeData(raw)
	if err != nil {
		value = serialization.NondecodedType(serialization.TypeToString(valueType))
	}
	row := output.Row{
		output.Column{
			Name:  output.NameValue,
			Type:  valueType,
			Value: value,
		},
	}
	if ec.Props().GetBool(mapFlagShowType) {
		row = append(row, output.Column{
			Name:  output.NameValueType,
			Type:  serialization.TypeString,
			Value: serialization.TypeToString(valueType),
		})
	}
	ec.AddOutputRows(row)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:get", &MapGetCommand{}))
}
