//go:build base || topic

package topic

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/topic"
)

type topicSubscribeCommand struct{}

func (mc *topicSubscribeCommand) Init(cc plug.InitContext) error {
	help := "Subscribe to a Topic for new messages."
	cc.SetCommandHelp(help, help)
	cc.AddIntFlag(topicFlagCount, "", 0, false, "number of messages to receive")
	cc.SetCommandUsage("subscribe")
	return nil
}

func (mc *topicSubscribeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	topicName := ec.Props().GetString(topicFlagName)
	tp, err := ec.Props().GetBlocking(topicPropertyName)
	if err != nil {
		return err
	}
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	t := tp.(*hazelcast.Topic)
	events := make(chan *topic.TopicEvent, 1)
	defer close(events)
	sid, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Listening for messages for topic %s", t.Name()))
		sid, err := topic.AddListener(ctx, ci, topicName, ec.Logger(), func(event *topic.TopicEvent) {
			// recover in case the channel is closed and there are unprocessed events
			defer recover()
			events <- event
		})
		if err != nil {
			return nil, err
		}
		return sid, nil
	})
	if err != nil {
		return err
	}
	defer topic.RemoveListener(ctx, ci, sid.(types.UUID))
	defer stop()
	return updateOutput(ctx, ec, events)
}

func init() {
	Must(plug.Registry.RegisterCommand("topic:subscribe", &topicSubscribeCommand{}))
}
