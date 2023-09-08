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

type MapRemoveCommand struct{}

func (mc *MapRemoveCommand) Unwrappable() {}

func (mc *MapRemoveCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("remove")
	help := "Remove a value from the given Map"
	cc.SetCommandHelp(help, help)
	addKeyTypeFlag(cc)
	cc.AddStringArg(argKey, argTitleKey)
	return nil
}

func (mc *MapRemoveCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(base.FlagName)
	showType := ec.Props().GetBool(base.FlagShowType)
	rowV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		keyStr := ec.GetStringArg(argKey)
		keyData, err := makeKeyData(ec, ci, keyStr)
		if err != nil {
			return nil, err
		}
		req := codec.EncodeMapRemoveRequest(mapName, keyData, 0)
		sp.SetText(fmt.Sprintf("Removing from map %s", mapName))
		resp, err := ci.InvokeOnKey(ctx, req, keyData, nil)
		if err != nil {
			return nil, err
		}
		raw := codec.DecodeMapRemoveResponse(resp)
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
	msg := fmt.Sprintf("OK Removed the entry from map: %s.\n", mapName)
	ec.PrintlnUnnecessary(msg)
	return ec.AddOutputRows(ctx, rowV.(output.Row))
}

func init() {
	Must(plug.Registry.RegisterCommand("map:remove", &MapRemoveCommand{}))
}
