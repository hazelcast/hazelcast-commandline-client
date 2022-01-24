package cobraprompt

import (
	"context"
	"time"
)

// CobraMutableCtx is a workaround for https://github.com/spf13/cobra/issues/1109
type CobraMutableCtx struct {
	Internal context.Context
}

func (c *CobraMutableCtx) Deadline() (deadline time.Time, ok bool) {
	return c.Internal.Deadline()
}
func (c *CobraMutableCtx) Done() <-chan struct{} {
	return c.Internal.Done()
}
func (c *CobraMutableCtx) Err() error {
	return c.Internal.Err()
}
func (c *CobraMutableCtx) Value(key interface{}) interface{} {
	return c.Internal.Value(key)
}
