//go:build std || topic

package topic

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
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
	cc.SetCommandUsage("topic")
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.SetTopLevel(true)
	help := "Topic operations"
	cc.SetCommandHelp(help, help)
	cc.AddStringFlag(topicFlagName, "n", defaultTopicName, false, "topic name")
	cc.AddBoolFlag(topicFlagShowType, "", false, false, "add the type names to the output")
	return nil
}

func (tc *TopicCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (tc *TopicCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	props.SetBlocking(topicPropertyName, func() (any, error) {
		topicName := ec.Props().GetString(topicFlagName)
		// empty topic name is allowed
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		tv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			sp.SetText(fmt.Sprintf("Getting topic %s", topicName))
			t, err := ci.Client().GetTopic(ctx, topicName)
			if err != nil {
				return nil, err
			}
			return t, nil
		})
		if err != nil {
			return nil, err
		}
		stop()
		return tv.(*hazelcast.Topic), nil
	})
	return nil
}

func init() {
	cmd := &TopicCommand{}
	Must(plug.Registry.RegisterCommand("topic", cmd))
	plug.Registry.RegisterAugmentor("20-topic", cmd)
}
