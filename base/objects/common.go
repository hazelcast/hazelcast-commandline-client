package objects

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
)

type Destroyer interface {
	Name() string
	Destroy(ctx context.Context) error
}

type getDestroyerFunc[T Destroyer] func(context.Context, plug.ExecContext, clc.Spinner) (T, error)

func CommandDestroyExec[T Destroyer](ctx context.Context, ec plug.ExecContext, typeName string, getFn getDestroyerFunc[T]) error {
	name := ec.Props().GetString(base.FlagName)
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo(fmt.Sprintf("%s '%s' will be deleted irreversibly, proceed?", typeName, name))
		if err != nil {
			ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		m, err := getFn(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Destroying %s '%s'", typeName, m.Name()))
		if err := m.Destroy(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Destroyed %s '%s'.", typeName, name)
	ec.PrintlnUnnecessary(msg)
	return nil
}
