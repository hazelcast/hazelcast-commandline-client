package _map

import (
	"context"
	"errors"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
)

const (
	mapDestroyFlagAutoYes = "yes"
)

type MapDestroyCommand struct{}

func (mc *MapDestroyCommand) Init(cc plug.InitContext) error {
	help := "Destroy a Map"
	cc.SetCommandHelp(help, help)
	cc.AddBoolFlag(mapDestroyFlagAutoYes, "", false, false, "skip interactive approval")
	cc.SetCommandUsage("destroy")
	return nil
}

func (mc *MapDestroyCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mv, err := ec.Props().GetBlocking(mapPropertyName)
	if err != nil {
		return err
	}

	autoYes := ec.Props().GetBool(mapDestroyFlagAutoYes)
	if !autoYes {
		prompt := prompt.NewPrompter(ec.Stdin(), ec.Stdout())
		yes, err := prompt.YesNoPrompt("Map will be deleted irreversibly, do you agree?")
		if err != nil {
			return err
		}
		if !yes {
			return errors.New("User did not agree")
		}
	}

	m := mv.(*hazelcast.Map)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Destroying map %s", m.Name()))
		if err := m.Destroy(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:destroy", &MapDestroyCommand{}))
}
