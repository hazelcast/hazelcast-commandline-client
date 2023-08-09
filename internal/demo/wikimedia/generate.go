package wikimedia

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fatih/structs"

	"github.com/hazelcast/hazelcast-commandline-client/internal/demo"
	"github.com/hazelcast/hazelcast-commandline-client/internal/sse"
)

const (
	streamURL = "https://stream.wikimedia.org/v2/stream/recentchange"
)

type StreamGenerator struct{}

func (StreamGenerator) Stream(ctx context.Context) chan demo.StreamItem {
	itemCh := make(chan demo.StreamItem, 1)
	client := sse.NewClient(streamURL)
	go func() {
		// retry logic
		for {
			err := handleEvents(ctx, client, itemCh)
			if err != nil {
				if err == sse.ErrNoConnection {
					break
				}
				if err == context.Canceled {
					break
				}
				// Retry all other errors including EOF
				time.Sleep(2 * time.Second)
				continue
			}
		}
		close(itemCh)
	}()
	return itemCh
}

func handleEvents(ctx context.Context, client *sse.Client, itemCh chan demo.StreamItem) error {
	evCh, errCh := client.Subscribe(ctx)
	for {
		select {
		case ev := <-evCh:
			if ev == nil {
				continue
			}
			it := &event{}
			err := json.Unmarshal(ev.Data, it)
			if err != nil {
				// XXX: should we log
				continue
			}
			select {
			case <-ctx.Done():
				// allows to exit from the loop when consumer exits without reading last item
				return ctx.Err()
			case itemCh <- it:
			}
		case err := <-errCh:
			return err
		}
	}
}

func (StreamGenerator) MappingQuery(mapName string) (string, error) {
	s := structs.New(flatEvent{})
	s.TagName = "json"
	return demo.GenerateMappingQuery(mapName, s.Map())
}
