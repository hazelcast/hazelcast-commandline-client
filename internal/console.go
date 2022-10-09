package internal

import (
	"context"
	"os"
	"os/signal"
)

func WaitTerminate(ctx context.Context, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	select {
	case <-c:
		cancel()
	case <-ctx.Done():
	}
}
