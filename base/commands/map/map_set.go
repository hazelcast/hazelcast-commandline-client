//go:build base || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type MapSetCommand struct{}

func (mc *MapSetCommand) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	addValueTypeFlag(cc)
	cc.AddIntFlag(mapTTL, "", ttlUnset, false, "time-to-live (ms)")
	cc.AddIntFlag(mapMaxIdle, "", ttlUnset, false, "max idle (ms)")
	cc.SetPositionalArgCount(2, 2)
	help := "Set a value in the given Map"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("set [-n MAP] KEY VALUE [flags]")
	return nil
}

func (mc *MapSetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(mapFlagName)
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
	ttl := GetTTL(ec)
	maxIdle := GetMaxIdle(ec)
	var req *hazelcast.ClientMessage
	if maxIdle >= 0 {
		req = codec.EncodeMapSetWithMaxIdleRequest(mapName, kd, vd, 0, ttl, maxIdle)
	} else {
		req = codec.EncodeMapSetRequest(mapName, kd, vd, 0, ttl)
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Setting value into map %s", mapName))
		return ci.InvokeOnKey(ctx, req, kd, nil)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:set", &MapSetCommand{}))
}
