package _map

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type MapPutCommand struct{}

func (mc *MapPutCommand) Init(cc plug.InitContext) error {
	cc.AddStringFlag(mapFlagKeyType, "k", "", false, "key type")
	cc.AddStringFlag(mapFlagValueType, "v", "", false, "value type")
	cc.AddIntFlag(mapTTL, "", ttlUnset, false, "time-to-live (ms)")
	cc.SetPositionalArgCount(2, 2)
	help := "Put a value to the given IMap and return the old value"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("put KEY VALUE")
	return nil
}

func (mc *MapPutCommand) Exec(ec plug.ExecContext) error {
	ctx := context.TODO()
	mapName := ec.Props().GetString(mapFlagName)
	ttl := GetTTL(ec)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	keyStr := ec.Args()[0]
	valueStr := ec.Args()[1]
	kd, vd, err := MakeKeyValueData(ec, ci, keyStr, valueStr)
	if err != nil {
		return err
	}
	req := codec.EncodeMapPutRequest(mapName, kd, vd, 0, ttl)
	resp, err := ci.InvokeOnKey(ctx, req, kd, nil)
	if err != nil {
		return err
	}
	raw := codec.DecodeMapPutResponse(resp)
	vt := raw.Type()
	value, err := ci.DecodeData(raw)
	if err != nil {
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
	Must(plug.Registry.RegisterCommand("map:put", &MapPutCommand{}))
}
