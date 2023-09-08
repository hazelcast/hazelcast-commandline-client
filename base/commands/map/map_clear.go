//go:build std || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type MapClearCommand struct{}

func (mc *MapClearCommand) Unwrappable() {}

func (mc *MapClearCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("clear")
	help := "Delete all entries of a Map"
	cc.SetCommandHelp(help, help)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the clear operation")
	return nil
}

func (mc *MapClearCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo("Map content will be deleted irreversibly, proceed?")
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
		sp.SetText(fmt.Sprintf("Clearing map %s", m.Name()))
		if err := m.Clear(ctx); err != nil {
			return nil, err
		}
		return m.Name(), nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Cleared map: %s.", name)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:clear", &MapClearCommand{}))
}
