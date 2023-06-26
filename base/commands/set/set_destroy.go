package set

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-go-client"
)

type SetDestroyCommand struct{}

func (qc *SetDestroyCommand) Init(cc plug.InitContext) error {
	long := `Destroy a Set
This command will delete the Set and the data in it will not be available anymore.`
	short := "Destroy a Set"
	cc.SetCommandHelp(long, short)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the destroy operation")
	cc.SetCommandUsage("destroy")
	return nil
}

func (qc *SetDestroyCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	sv, err := ec.Props().GetBlocking(setPropertyName)
	if err != nil {
		return err
	}
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo("Set will be deleted irreversibly, proceed?")
		if err != nil {
			ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	s := sv.(*hazelcast.Set)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Destroying set %s", s.Name()))
		if err := s.Destroy(ctx); err != nil {
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
	Must(plug.Registry.RegisterCommand("set:destroy", &SetDestroyCommand{}))
}
