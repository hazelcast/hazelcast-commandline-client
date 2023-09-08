//go:build std || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type MapUnlock struct{}

func (mc *MapUnlock) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("unlock")
	long := `Unlock a key in the given Map

This command is only available in the interactive mode.`
	short := "Unlock a key in the given Map"
	cc.SetCommandHelp(long, short)
	addKeyTypeFlag(cc)
	cc.AddStringArg(argKey, argTitleKey)
	return nil
}

func (mc *MapUnlock) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(base.FlagName)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Unlocking key in map %s", mapName))
		m, err := getMap(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		keyStr := ec.GetStringArg(argKey)
		keyData, err := makeKeyData(ec, ci, keyStr)
		if err != nil {
			return nil, err
		}
		return nil, m.Unlock(ctx, keyData)
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Unlocked the key in map %s", mapName)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:unlock", &MapUnlock{}, plug.OnlyInteractive{}))
}
