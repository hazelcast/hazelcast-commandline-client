//go:build std || map

package _map

import (
	"context"
	"fmt"
	"math"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type MapLoadAllCommand struct{}

func (mc *MapLoadAllCommand) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	help := "Load keys from map-store into the map. If no key is given, all keys are loaded."
	cc.AddBoolFlag(mapFlagReplace, "", false, false, "replace keys if they exist in the map")
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("load-all [keys] [flags]")
	cc.SetPositionalArgCount(0, math.MaxInt)
	return nil
}

func (mc *MapLoadAllCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(mapFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	var keys []hazelcast.Data
	for _, keyStr := range ec.Args() {
		keyData, err := makeKeyData(ec, ci, keyStr)
		if err != nil {
			return err
		}
		keys = append(keys, keyData)
	}
	replace := ec.Props().GetBool(mapFlagReplace)
	var req *hazelcast.ClientMessage
	if len(keys) == 0 {
		req = codec.EncodeMapLoadAllRequest(mapName, replace)
	} else {
		req = codec.EncodeMapLoadGivenKeysRequest(mapName, keys, replace)
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Loading keys into the map %s", mapName))
		return ci.InvokeOnRandomTarget(ctx, req, nil)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:load-all", &MapLoadAllCommand{}))
}
