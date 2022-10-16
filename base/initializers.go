package base

import (
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/groups"
	"github.com/hazelcast/hazelcast-commandline-client/clc/property"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type GlobalInitializer struct{}

func (g GlobalInitializer) Init(cc plug.InitContext) error {
	// base group IDs
	cc.AddCommandGroup(groups.DDSID, "Distributed Data Structures")
	// output-type flag
	pns := plug.Registry.PrinterNames()
	usage := fmt.Sprintf("set the output type, one of: %s", strings.Join(pns, ", "))
	// other flags
	cc.AddStringFlag(property.OutputType, "", "", false, usage)
	cc.AddBoolFlag(property.Verbose, "", false, false, "enable verbose output")
	if !cc.Interactive() {
		cc.AddStringFlag(property.ConfigPath, "c", "", false, "set the configuration path")
		cc.AddStringFlag(property.ClusterAddress, "a", "", false, "set the cluster address")
	}
	// configuration
	cc.AddStringConfig(property.ClusterAddress, "localhost:5701", property.ClusterAddress, "cluster address")
	return nil
}

func init() {
	plug.Registry.RegisterGlobalInitializer("00-global-init", &GlobalInitializer{})
}
