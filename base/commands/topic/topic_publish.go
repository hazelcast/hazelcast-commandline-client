//go:build std || topic

package topic

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type PublishCommand struct{}

func (PublishCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("publish")
	help := "Publish new messages for a Topic."
	cc.SetCommandHelp(help, help)
	commands.AddValueTypeFlag(cc)
	cc.AddStringSliceArg(base.ArgValue, base.ArgTitleValue, 1, clc.MaxArgs)
	return nil
}

func (PublishCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	vt := ec.Props().GetString(base.FlagValueType)
	countV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		cid, vid := cmd.FindClusterIDs(ctx, ec)
		ec.Metrics().Increment(metrics.NewKey(cid, vid), "total.topic."+cmd.RunningModeString(ec))
		// get the topic just to ensure the corresponding proxy is created
		t, err := ci.Client().GetTopic(ctx, name)
		if err != nil {
			return nil, err
		}
		args := ec.GetStringSliceArg(base.ArgValue)
		vs := make([]any, len(args))
		for i, arg := range args {
			val, err := mk.ValueFromString(arg, vt)
			if err != nil {
				return nil, err
			}
			vs[i] = val
		}
		sp.SetText(fmt.Sprintf("Publishing %d values to Topic '%s'", len(vs), name))
		if err := t.PublishAll(ctx, vs...); err != nil {
			return nil, fmt.Errorf("publishing values: %w", err)
		}
		return len(vs), nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Published %d values to Topic '%s'.", countV.(int), name)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("topic:publish", &PublishCommand{}))
}
