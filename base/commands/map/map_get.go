//go:build std || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type MapGetCommand struct{}

func (mc *MapGetCommand) Unwrappable() {}

func (mc *MapGetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("get")
	addKeyTypeFlag(cc)
	help := "Get a value from the given Map"
	cc.SetCommandHelp(help, help)
	cc.AddStringArg(argKey, argTitleKey)
	return nil
}

func (mc *MapGetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(base.FlagName)
	keyStr := ec.GetStringArg(argKey)
	showType := ec.Props().GetBool(base.FlagShowType)
	rowV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Getting from map %s", mapName))
		keyData, err := makeKeyData(ec, ci, keyStr)
		if err != nil {
			return nil, err
		}
		req := codec.EncodeMapGetRequest(mapName, keyData, 0)
		resp, err := ci.InvokeOnKey(ctx, req, keyData, nil)
		if err != nil {
			return nil, err
		}
		data := codec.DecodeMapGetResponse(resp)
		vt := data.Type()
		value, err := ci.DecodeData(data)
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
		if showType {
			row = append(row, output.Column{
				Name:  output.NameValueType,
				Type:  serialization.TypeString,
				Value: serialization.TypeToLabel(vt),
			})
		}
		return row, nil
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, rowV.(output.Row))
}

func init() {
	Must(plug.Registry.RegisterCommand("map:get", &MapGetCommand{}))
}
