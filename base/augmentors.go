package base

import (
	"context"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	PropertyClientInternalName = "ci"
)

type ClientInternalAugmentor struct {
	ci *hazelcast.ClientInternal
}

func (c *ClientInternalAugmentor) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	props.SetBlocking(PropertyClientInternalName, func() (any, error) {
		if c.ci != nil {
			return c.ci, nil
		}
		client, err := ec.Client(ctx)
		if err != nil {
			return nil, err
		}
		ci := hazelcast.NewClientInternal(client)
		c.ci = ci
		return ci, nil
	})
	return nil
}

func init() {
	plug.Registry.RegisterAugmentor("00-client-internal", &ClientInternalAugmentor{})
}
