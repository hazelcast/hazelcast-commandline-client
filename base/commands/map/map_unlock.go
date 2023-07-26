//go:build base || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
)

type MapUnlock struct{}

func (mc *MapUnlock) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	long := `Unlock a key in the given Map

This command is only available in the interactive mode.`
	short := "Unlock a key in the given Map"
	cc.SetCommandHelp(long, short)
	cc.SetCommandUsage("unlock [key] [flags]")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (mc *MapUnlock) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(mapFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	mv, err := ec.Props().GetBlocking(mapPropertyName)
	if err != nil {
		return err
	}
	m := mv.(*hazelcast.Map)
	keyStr := ec.Args()[0]
	keyData, err := makeKeyData(ec, ci, keyStr)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Locking key in map %s", mapName))
		return nil, m.Unlock(ctx, keyData)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:unlock", &MapUnlock{}, plug.OnlyInteractive{}))
}
