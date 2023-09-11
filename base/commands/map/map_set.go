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

type MapSetCommand struct{}

func (MapSetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("set")
	help := "Set a value in the given Map"
	cc.SetCommandHelp(help, help)
	commands.AddKeyTypeFlag(cc)
	commands.AddValueTypeFlag(cc)
	cc.AddIntFlag(commands.FlagTTL, "", clc.TTLUnset, false, "time-to-live (ms)")
	cc.AddIntFlag(mapMaxIdle, "", clc.TTLUnset, false, "max idle (ms)")
	cc.AddStringArg(commands.ArgKey, commands.ArgTitleKey)
	cc.AddStringArg(base.ArgValue, base.ArgTitleValue)
	return nil
}

func (MapSetCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(base.FlagName)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Setting value into Map '%s'", mapName))
		_, err = getMap(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		key := ec.GetStringArg(commands.ArgKey)
		value := ec.GetStringArg(base.ArgValue)
		kd, vd, err := commands.MakeKeyValueData(ec, ci, key, value)
		if err != nil {
			return nil, err
		}
		ttl := commands.GetTTL(ec)
		maxIdle := getMaxIdle(ec)
		var req *hazelcast.ClientMessage
		if maxIdle >= 0 {
			req = codec.EncodeMapSetWithMaxIdleRequest(mapName, kd, vd, 0, ttl, maxIdle)
		} else {
			req = codec.EncodeMapSetRequest(mapName, kd, vd, 0, ttl)
		}
		if _, err = ci.InvokeOnKey(ctx, req, kd, nil); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Set the value into the Map '%s'", mapName)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:set", &MapSetCommand{}))
}
