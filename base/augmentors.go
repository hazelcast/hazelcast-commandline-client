package base

import (
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type CheckOutputTypeAugmentor struct{}

func (ag *CheckOutputTypeAugmentor) Augment(ec plug.ExecContext, props *plug.Properties) error {
	pns := map[string]struct{}{}
	for _, n := range plug.Registry.PrinterNames() {
		pns[n] = struct{}{}
	}
	ot := ec.Props().GetString(clc.PropertyFormat)
	if ot == "" {
		props.Set(clc.PropertyFormat, "delimited")
		return nil
	}
	if _, ok := pns[ot]; !ok {
		return fmt.Errorf("invalid %s: %s", clc.PropertyFormat, ot)
	}
	return nil
}

func init() {
	plug.Registry.RegisterAugmentor("00-check-output-type", &CheckOutputTypeAugmentor{})
}
