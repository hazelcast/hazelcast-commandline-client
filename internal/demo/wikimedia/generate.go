package wikimedia

import (
	"context"
	"encoding/json"

	"github.com/fatih/structs"
	"github.com/r3labs/sse"

	"github.com/hazelcast/hazelcast-commandline-client/internal/demo"
)

const (
	streamURL = "https://stream.wikimedia.org/v2/stream/recentchange"
)

type StreamGenerator struct{}

func (StreamGenerator) Stream(ctx context.Context) chan demo.StreamItem {
	eventCh := make(chan *sse.Event)
	itemCh := make(chan demo.StreamItem)
	client := sse.NewClient(streamURL)
	client.SubscribeChanWithContext(ctx, "messages", eventCh)
	go func() {
		for {
			select {
			case ev := <-eventCh:
				it := &event{}
				err := json.Unmarshal(ev.Data, it)
				if err != nil {
					// XXX: should we log
					continue
				}
				itemCh <- it
			case <-ctx.Done():
				return
			}
		}
	}()
	return itemCh
}

func (StreamGenerator) MappingQuery(mapName string) (string, error) {
	s := structs.New(flatEvent{})
	s.TagName = "json"
	return demo.GenerateMappingQuery(mapName, s.Map())
}
