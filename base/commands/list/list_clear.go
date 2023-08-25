//go:build std || list

package list

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type ListClearCommand struct{}

func (mc *ListClearCommand) Init(cc plug.InitContext) error {
	help := "Delete all entries of a List"
	cc.SetCommandHelp(help, help)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the clear operation")
	cc.SetCommandUsage("clear [flags]")
	return nil
}

func (mc *ListClearCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	lv, err := ec.Props().GetBlocking(listPropertyName)
	if err != nil {
		return err
	}
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo("List content will be deleted irreversibly, proceed?")
		if err != nil {
			ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	l := lv.(*hazelcast.List)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Clearing list %s", l.Name()))
		if err := l.Clear(ctx); err != nil {
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
	Must(plug.Registry.RegisterCommand("list:clear", &ListClearCommand{}))
}
