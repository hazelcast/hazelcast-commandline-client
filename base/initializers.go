package base

import (
	"fmt"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/groups"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
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
	lp := paths.DefaultLogPath(time.Now())
	cc.AddStringFlag(property.LogFile, "", lp, false, "set the log file, use stderr to log to stderr")
	cc.AddStringFlag(property.LogLevel, "", "info", false, "set the log level")
	cc.AddStringFlag(property.SchemaDir, "", "", false, "set the schema directory")
	// configuration
	cc.AddStringConfig(property.ClusterAddress, "localhost:5701", property.ClusterAddress, "cluster address")
	cc.AddStringConfig(property.LogFile, "", property.LogFile, "log file")
	cc.AddStringConfig(property.LogLevel, "", property.LogLevel, "log level")
	cc.AddStringConfig(property.SchemaDir, "", property.SchemaDir, "schema directory")
	return nil
}

func init() {
	plug.Registry.RegisterGlobalInitializer("00-global-init", &GlobalInitializer{})
}
