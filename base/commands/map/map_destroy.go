//go:build std || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
)

type MapDestroyCommand struct{}

func (mc *MapDestroyCommand) Unwrappable() {}

func (mc *MapDestroyCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("destroy")
	long := `Destroy a Map

This command will delete the Map and the data in it will not be available anymore.`
	short := "Destroy a Map"
	cc.SetCommandHelp(long, short)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the destroy operation")
	return nil
}

func (mc *MapDestroyCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo("Map will be deleted irreversibly, proceed?")
		if err != nil {
			ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	name, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		m, err := getMap(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Destroying map %s", m.Name()))
		if err := m.Destroy(ctx); err != nil {
			return nil, err
		}
		return m.Name(), nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Destroyed map %s.", name)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:destroy", &MapDestroyCommand{}))
}
