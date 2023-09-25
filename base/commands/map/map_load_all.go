//go:build std || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type MapLoadAllCommand struct{}

func (MapLoadAllCommand) Init(cc plug.InitContext) error {
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

func (MapLoadAllCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		replace := ec.Props().GetBool(mapFlagReplace)
		keyStrs := ec.GetStringSliceArg(commands.ArgKey)
		keys := make([]any, len(keyStrs))
		kt := ec.Props().GetString(commands.FlagKeyType)
		keys, err := commands.MakeValuesFromStrings(kt, keyStrs)
		if err != nil {
			return nil, err
		}
		m, err := getMap(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		if replace {
			return nil, m.LoadAllReplacing(ctx, keys...)
		}
		return nil, m.LoadAllWithoutReplacing(ctx, keys...)
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
