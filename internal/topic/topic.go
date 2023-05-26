package topic

import (
	"context"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type TopicEvent struct {
	PublishTime time.Time
	Value       interface{}
	ValueType   int32
	TopicName   string
	Member      cluster.MemberInfo
}

func newTopicEvent(name string, value interface{}, valueType int32, publishTime time.Time, member cluster.MemberInfo) *TopicEvent {
	return &TopicEvent{
		TopicName:   name,
		Value:       value,
		ValueType:   valueType,
		PublishTime: publishTime,
		Member:      member,
	}
}

func PublishAll(ctx context.Context, ci *hazelcast.ClientInternal, tName string, vals []hazelcast.Data) error {
	pid, err := stringToPartitionID(ci, tName)
	if err != nil {
		return err
	}
	req := codec.EncodeTopicPublishAllRequest(tName, vals)
	_, err = ci.InvokeOnPartition(ctx, req, pid, nil)
	return err
}

func stringToPartitionID(ci *hazelcast.ClientInternal, name string) (int32, error) {
	idx := strings.Index(name, "@")
	if keyData, err := ci.EncodeData(name[idx+1:]); err != nil {
		return 0, err
	} else if partitionID, err := ci.GetPartitionID(keyData); err != nil {
		return 0, err
	} else {
		return partitionID, nil
	}
}

func AddListener(ctx context.Context, ci *hazelcast.ClientInternal, tName string, logger log.Logger, handler func(event *TopicEvent)) (types.UUID, error) {
	subscriptionID := types.NewUUID()
	addRequest := codec.EncodeTopicAddMessageListenerRequest(tName, false)
	removeRequest := codec.EncodeTopicRemoveMessageListenerRequest(tName, subscriptionID)
	listenerHandler := func(msg *hazelcast.ClientMessage) {
		codec.HandleTopicAddMessageListener(msg, func(itemData hazelcast.Data, publishTime int64, uuid types.UUID) {
			item, err := ci.DecodeData(itemData)
			if err != nil {
				logger.Warn("The value was not decoded, due to error: %s", err.Error())
				return
			}
			itemType := itemData.Type()
			var member cluster.MemberInfo
			if m := ci.ClusterService().GetMemberByUUID(uuid); m != nil {
				member = *m
			}
			handler(newTopicEvent(tName, item, itemType, time.Unix(0, publishTime*1_000_000), member))
		})
	}
	binder := ci.ListenerBinder()
	err := binder.Add(ctx, subscriptionID, addRequest, removeRequest, listenerHandler)
	return subscriptionID, err
}

// RemoveListener removes the given subscription from this topic.
func RemoveListener(ctx context.Context, ci *hazelcast.ClientInternal, subscriptionID types.UUID) error {
	return ci.ListenerBinder().Remove(ctx, subscriptionID)
}
