//go:build std || list

package list

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
)

type ListDestroyCommand struct{}

func (mc *ListDestroyCommand) Unwrappable() {}

func (mc *ListDestroyCommand) Init(cc plug.InitContext) error {
	long := `Destroy a List

This command will delete the List and the data in it will not be available anymore.`
	cc.SetCommandUsage("destroy")
	short := "Destroy a List"
	cc.SetCommandHelp(long, short)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the destroy operation")
	return nil
}

func (mc *ListDestroyCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo("List will be deleted irreversibly, proceed?")
		if err != nil {
			ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	name, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		l, err := getList(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Destroying list %s", l.Name()))
		if err := l.Destroy(ctx); err != nil {
			return nil, err
		}
		return l.Name(), nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Destroyed list %s.", name)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("list:destroy", &ListDestroyCommand{}))
}
