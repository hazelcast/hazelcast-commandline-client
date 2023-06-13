//go:build base || topic

package topic

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	topicFlagName     = "name"
	topicFlagShowType = "show-type"
	topicPropertyName = "topic"
)

type TopicCommand struct {
}

func (mc *TopicCommand) Init(cc plug.InitContext) error {
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.AddStringFlag(topicFlagName, "n", defaultTopicName, false, "topic name")
	cc.AddBoolFlag(topicFlagShowType, "", false, false, "add the type names to the output")
	if !cc.Interactive() {
		cc.AddStringFlag(clc.PropertySchemaDir, "", paths.Schemas(), false, "set the schema directory")
	}
	cc.SetTopLevel(true)
	cc.SetCommandUsage("topic [command] [flags]")
	help := "Topic operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *TopicCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (mc *TopicCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	props.SetBlocking(topicPropertyName, func() (any, error) {
		topicName := ec.Props().GetString(topicFlagName)
		// empty topic name is allowed
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		mv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			sp.SetText(fmt.Sprintf("Getting topic %s", topicName))
			m, err := ci.Client().GetTopic(ctx, topicName)
			if err != nil {
				return nil, err
			}
			return m, nil
		})
		if err != nil {
			return nil, err
		}
		stop()
		return mv.(*hazelcast.Topic), nil
	})
	return nil
}

func init() {
	cmd := &TopicCommand{}
	Must(plug.Registry.RegisterCommand("topic", cmd))
	plug.Registry.RegisterAugmentor("20-topic", cmd)
}
