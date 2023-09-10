//go:build std || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type MapLoadAllCommand struct{}

func (mc *MapLoadAllCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("load-all")
	long := `Load keys from map-store into the map
	
If no key is given, all keys are loaded.`
	short := "Load keys from map-store into the map"
	cc.SetCommandHelp(long, short)
	commands.AddKeyTypeFlag(cc)
	cc.AddBoolFlag(mapFlagReplace, "", false, false, "replace keys if they exist in the map")
	cc.AddStringSliceArg(commands.ArgKey, commands.ArgTitleKey, 0, clc.MaxArgs)
	return nil
}

func (mc *MapLoadAllCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		var keys []hazelcast.Data
		for _, keyStr := range ec.GetStringSliceArg(commands.ArgKey) {
			keyData, err := commands.MakeKeyData(ec, ci, keyStr)
			if err != nil {
				return nil, err
			}
			keys = append(keys, keyData)
		}
		replace := ec.Props().GetBool(mapFlagReplace)
		var req *hazelcast.ClientMessage
		if len(keys) == 0 {
			req = codec.EncodeMapLoadAllRequest(name, replace)
		} else {
			req = codec.EncodeMapLoadGivenKeysRequest(name, keys, replace)
		}
		sp.SetText(fmt.Sprintf("Loading keys into the Map '%s'", name))
		if _, err = ci.InvokeOnRandomTarget(ctx, req, nil); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Loaded the keys into Map '%s'", name)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:load-all", &MapLoadAllCommand{}))
}
