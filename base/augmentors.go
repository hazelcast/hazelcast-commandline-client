package base

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc/property"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ClientInternalAugmentor struct {
	ci *hazelcast.ClientInternal
}

func (c *ClientInternalAugmentor) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	props.SetBlocking(property.ClientInternal, func() (any, error) {
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

type CheckOutputTypeAugmentor struct{}

func (ag *CheckOutputTypeAugmentor) Augment(ec plug.ExecContext, props *plug.Properties) error {
	pns := map[string]struct{}{}
	for _, n := range plug.Registry.PrinterNames() {
		pns[n] = struct{}{}
	}
	ot := ec.Props().GetString(property.OutputType)
	if ot == "" {
		props.Set(property.OutputType, "delimited")
		return nil
	}
	_, ok := pns[ot]
	if !ok {
		return fmt.Errorf("invalid %s: %s", property.OutputType, ot)
	}
	return nil
}

func init() {
	plug.Registry.RegisterAugmentor("00-client-internal", &ClientInternalAugmentor{})
	plug.Registry.RegisterAugmentor("00-check-output-type", &CheckOutputTypeAugmentor{})
}
