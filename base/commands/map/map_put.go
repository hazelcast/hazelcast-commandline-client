//go:build base || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type MapPutCommand struct{}

func (mc *MapPutCommand) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	addValueTypeFlag(cc)
	cc.AddIntFlag(mapTTL, "", ttlUnset, false, "time-to-live (ms)")
	cc.SetPositionalArgCount(2, 2)
	help := "Put a value in the given Map and return the old value"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("put [-n MAP] KEY VALUE [flags]")
	return nil
}

func (mc *MapPutCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(mapFlagName)
	ttl := GetTTL(ec)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	// get the map just to ensure the corresponding proxy is created
	if _, err := ec.Props().GetBlocking(mapPropertyName); err != nil {
		return err
	}
	keyStr := ec.Args()[0]
	valueStr := ec.Args()[1]
	kd, vd, err := makeKeyValueData(ec, ci, keyStr, valueStr)
	if err != nil {
		return err
	}
	req := codec.EncodeMapPutRequest(mapName, kd, vd, 0, ttl)
	hint := fmt.Sprintf("Putting into map %s", mapName)
	rv, stop, err := ec.ExecuteBlocking(ctx, hint, func(ctx context.Context) (any, error) {
		return ci.InvokeOnKey(ctx, req, kd, nil)
	})
	if err != nil {
		return err
	}
	stop()
	raw := codec.DecodeMapPutResponse(rv.(*hazelcast.ClientMessage))
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
	return ec.AddOutputRows(ctx, row)
}

func init() {
	Must(plug.Registry.RegisterCommand("map:put", &MapPutCommand{}))
}
