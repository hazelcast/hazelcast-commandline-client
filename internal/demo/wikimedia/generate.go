package wikimedia

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/internal/demo"
	"github.com/hazelcast/hazelcast-commandline-client/internal/sse"
)

const (
	streamURL = "https://stream.wikimedia.org/v2/stream/recentchange"
)

type StreamGenerator struct{}

func (StreamGenerator) Stream(ctx context.Context) (chan demo.StreamItem, context.CancelFunc) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	itemCh := make(chan demo.StreamItem)
	client := sse.NewClient(streamURL)
	go func() {
		// retry logic
		for {
			err := handleEvents(ctx, client, itemCh)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					break
				}
				if errors.Is(err, sse.ErrNoConnection) {
					break
				}
				// Retry all other errors including EOF
				time.Sleep(2 * time.Second)
				continue
			}
		}
		close(itemCh)
		cancel()
	}()
	return itemCh, cancel
}

func handleEvents(ctx context.Context, client *sse.Client, itemCh chan demo.StreamItem) error {
	return client.SubscribeWithCallback(ctx, func(rawEv *sse.Event) error {
		if rawEv == nil {
			return nil
		}
		ev := event{}
		err := json.Unmarshal(rawEv.Data, &ev)
		if err != nil {
			// XXX: should we log
			return nil
		}
		select {
		case itemCh <- ev:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	})
}

func (StreamGenerator) GenerateMappingQuery(mapName string) (string, error) {
	return demo.GenerateMappingQuery(mapName, event{}.KeyValues())
}
