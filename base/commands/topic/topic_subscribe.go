//go:build std || topic

package topic

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type SubscribeCommand struct{}

func (SubscribeCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("subscribe")
	help := "Subscribe to a Topic for new messages."
	cc.SetCommandHelp(help, help)
	cc.AddIntFlag(topicFlagCount, "", 0, false, "number of messages to receive")
	return nil
}

func (SubscribeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	events := make(chan TopicEvent, 1)
	// Channel is not closed intentionally
	sid, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Listening to messages of topic %s", name))
		sid, err := addListener(ctx, ci, name, ec.Logger(), func(event TopicEvent) {
			select {
			case events <- event:
			case <-ctx.Done():
			}
		})
		if err != nil {
			return nil, err
		}
		return sid, nil
	})
	if err != nil {
		return err
	}
	defer removeListener(ctx, ci, sid.(types.UUID))
	defer stop()
	return updateOutput(ctx, ec, events)
}

type TopicEvent struct {
	PublishTime time.Time
	Value       any
	ValueType   int32
	TopicName   string
	Member      cluster.MemberInfo
}

func newTopicEvent(name string, value any, valueType int32, publishTime time.Time, member cluster.MemberInfo) TopicEvent {
	return TopicEvent{
		TopicName:   name,
		Value:       value,
		ValueType:   valueType,
		PublishTime: publishTime,
		Member:      member,
	}
}

func addListener(ctx context.Context, ci *hazelcast.ClientInternal, topic string, logger log.Logger, handler func(event TopicEvent)) (types.UUID, error) {
	subscriptionID := types.NewUUID()
	addRequest := codec.EncodeTopicAddMessageListenerRequest(topic, false)
	removeRequest := codec.EncodeTopicRemoveMessageListenerRequest(topic, subscriptionID)
	listenerHandler := func(msg *hazelcast.ClientMessage) {
		codec.HandleTopicAddMessageListener(msg, func(itemData hazelcast.Data, publishTime int64, uuid types.UUID) {
			itemType := itemData.Type()
			item, err := ci.DecodeData(itemData)
			if err != nil {
				logger.Warn("The value was not decoded, due to error: %s", err.Error())
				item = serialization.NondecodedType(serialization.TypeToLabel(itemType))
			}
			var member cluster.MemberInfo
			if m := ci.ClusterService().GetMemberByUUID(uuid); m != nil {
				member = *m
			}
			handler(newTopicEvent(topic, item, itemType, time.Unix(0, publishTime*1_000_000), member))
		})
	}
	binder := ci.ListenerBinder()
	err := binder.Add(ctx, subscriptionID, addRequest, removeRequest, listenerHandler)
	return subscriptionID, err
}

// removeListener removes the given subscription from this topic.
func removeListener(ctx context.Context, ci *hazelcast.ClientInternal, subscriptionID types.UUID) error {
	return ci.ListenerBinder().Remove(ctx, subscriptionID)
}

func updateOutput(ctx context.Context, ec plug.ExecContext, events <-chan TopicEvent) error {
	wantedCount := int(ec.Props().GetInt(topicFlagCount))
	rowCh := make(chan output.Row)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()
	name := ec.Props().GetString(base.FlagName)
	ec.PrintlnUnnecessary(fmt.Sprintf("Listening to messages of Topic '%s'", name))
	go retrieveMessages(ctx, ec, wantedCount, events, rowCh)
	return ec.AddOutputStream(ctx, rowCh)
}

func retrieveMessages(ctx context.Context, ec plug.ExecContext, wanted int, events <-chan TopicEvent, rowCh chan<- output.Row) {
	printed := 0
loop:
	for {
		var e TopicEvent
		select {
		case e = <-events:
		case <-ctx.Done():
			break loop
		}
		row := eventRow(e, ec)
		select {
		case rowCh <- row:
		case <-ctx.Done():
			break loop
		}
		printed++
		if wanted > 0 && printed == wanted {
			break loop
		}
	}
	close(rowCh)
}

func eventRow(e TopicEvent, ec plug.ExecContext) (row output.Row) {
	if ec.Props().GetBool(clc.PropertyVerbose) {
		row = append(row,
			output.Column{
				Name:  "Time",
				Type:  serialization.TypeJavaLocalDateTime,
				Value: e.PublishTime,
			},
			output.Column{
				Name:  "Topic",
				Type:  serialization.TypeString,
				Value: e.TopicName,
			},
			output.Column{
				Name:  "Value",
				Type:  e.ValueType,
				Value: e.Value,
			},
		)
		if ec.Props().GetBool(base.FlagShowType) {
			row = append(row,
				output.Column{
					Name:  "Type",
					Type:  serialization.TypeString,
					Value: serialization.TypeToLabel(e.ValueType),
				})
		}
		row = append(row,
			output.Column{
				Name:  "Member UUID",
				Type:  serialization.TypeUUID,
				Value: e.Member.UUID,
			},
			output.Column{
				Name:  "Member Address",
				Type:  serialization.TypeString,
				Value: string(e.Member.Address),
			})
		return row
	}
	row = output.Row{
		output.Column{
			Name:  "Value",
			Type:  e.ValueType,
			Value: e.Value,
		},
	}
	if ec.Props().GetBool(base.FlagShowType) {
		row = append(row,
			output.Column{
				Name:  "Type",
				Type:  serialization.TypeString,
				Value: serialization.TypeToLabel(e.ValueType),
			})
	}
	return row
}

func init() {
	Must(plug.Registry.RegisterCommand("topic:subscribe", &SubscribeCommand{}))
}
