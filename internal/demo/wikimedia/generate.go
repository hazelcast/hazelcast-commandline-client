package wikimedia

import (
	"context"
	"encoding/json"
	"time"

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
	rawEventCh, errCh := client.Subscribe(ctx)
	for {
		select {
		case rawEv := <-rawEventCh:
			if rawEv == nil {
				continue
			}
			ev := event{}
			err := json.Unmarshal(rawEv.Data, &ev)
			if err != nil {
				// XXX: should we log
				continue
			}
			select {
			case <-ctx.Done():
				// allows to exit from the loop when consumer exits without reading last item
				return ctx.Err()
			case itemCh <- ev:
			}
		case err := <-errCh:
			return err
		}
	}
}

func (StreamGenerator) MappingQuery(mapName string) (string, error) {
	return demo.GenerateMappingQuery(mapName, event{}.KeyValues())
}
