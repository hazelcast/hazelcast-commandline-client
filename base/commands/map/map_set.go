//go:build std || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

const (
	argValue      = "value"
	argTitleValue = "value"
)

type MapSetCommand struct{}

func (mc *MapSetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("set")
	help := "Set a value in the given Map"
	cc.SetCommandHelp(help, help)
	addKeyTypeFlag(cc)
	addValueTypeFlag(cc)
	cc.AddIntFlag(mapTTL, "", ttlUnset, false, "time-to-live (ms)")
	cc.AddIntFlag(mapMaxIdle, "", ttlUnset, false, "max idle (ms)")
	cc.AddStringArg(argKey, argTitleKey)
	cc.AddStringArg(argValue, argTitleValue)
	return nil
}

func (mc *MapSetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(base.FlagName)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Setting value into map %s", mapName))
		_, err = getMap(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		keyStr := ec.GetStringArg(argKey)
		valueStr := ec.GetStringArg(argValue)
		kd, vd, err := makeKeyValueData(ec, ci, keyStr, valueStr)
		if err != nil {
			return nil, err
		}
		ttl := GetTTL(ec)
		maxIdle := GetMaxIdle(ec)
		var req *hazelcast.ClientMessage
		if maxIdle >= 0 {
			req = codec.EncodeMapSetWithMaxIdleRequest(mapName, kd, vd, 0, ttl, maxIdle)
		} else {
			req = codec.EncodeMapSetRequest(mapName, kd, vd, 0, ttl)
		}
		_, err = ci.InvokeOnKey(ctx, req, kd, nil)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Set the value into the map %s", mapName)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:set", &MapSetCommand{}))
}
