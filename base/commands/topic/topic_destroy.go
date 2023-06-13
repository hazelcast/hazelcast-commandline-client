//go:build base || topic

package topic

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-go-client"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type topicDestroyCommand struct{}

func (mc *topicDestroyCommand) Init(cc plug.InitContext) error {
	long := `Destroy a Topic

This command will delete the Topic and the data in it will not be available anymore.`
	short := "Destroy a Topic"
	cc.SetCommandHelp(long, short)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the destroy operation")
	cc.SetCommandUsage("destroy [flags]")
	return nil
}

func (mc *topicDestroyCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	tp, err := ec.Props().GetBlocking(topicPropertyName)
	if err != nil {
		return err
	}
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo("Topic will be deleted irreversibly, proceed?")
		if err != nil {
			ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	t := tp.(*hazelcast.Topic)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Destroying topic %s", t.Name()))
		err := t.Destroy(ctx)
		if err != nil {
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
	Must(plug.Registry.RegisterCommand("topic:destroy", &topicDestroyCommand{}))
}
