package base

import (
	"fmt"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/groups"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type GlobalInitializer struct{}

func (g GlobalInitializer) Init(cc plug.InitContext) error {
	// base group IDs
	cc.AddCommandGroup(groups.DDSID, "Distributed Data Structures")
	// output type flag
	pns := plug.Registry.PrinterNames()
	usage := fmt.Sprintf("set the output type, one of: %s", strings.Join(pns, ", "))
	// other flags
	cc.AddStringFlag(clc.PropertyOutputFormat, "", "", false, usage)
	cc.AddBoolFlag(clc.PropertyVerbose, "", false, false, "enable verbose output")
	lp := paths.DefaultLogPath(time.Now())
	if !cc.Interactive() {
		cc.AddStringFlag(clc.PropertyConfigPath, clc.ShortcutConfigPath, "", false, "set the configuration path")
		cc.AddStringFlag(clc.PropertyClusterAddress, "a", "", false, "set the cluster address")
		cc.AddStringFlag(clc.PropertyClusterName, "", "", false, "set the cluster name")
		cc.AddStringFlag(clc.PropertyLogPath, "", lp, false, "set the log path, use stderr to log to stderr")
		cc.AddStringFlag(clc.PropertyLogLevel, "", "info", false, "set the log level")
		cc.AddStringFlag(clc.PropertySchemaDir, "", "", false, "set the schema directory")
	}
	// configuration
	cc.AddStringConfig(clc.PropertyClusterAddress, "localhost:5701", clc.PropertyClusterAddress, "cluster address")
	cc.AddStringConfig(clc.PropertyClusterName, "dev", clc.PropertyClusterName, "cluster name")
	cc.AddStringConfig(clc.PropertyLogPath, "", clc.PropertyLogPath, "log path")
	cc.AddStringConfig(clc.PropertyLogLevel, "", clc.PropertyLogLevel, "log level")
	cc.AddStringConfig(clc.PropertySchemaDir, "", clc.PropertySchemaDir, "schema directory")
	cc.AddStringConfig(clc.PropertyViridianToken, "", "", "Viridian token")
	return nil
}

func init() {
	plug.Registry.RegisterGlobalInitializer("00-global-init", &GlobalInitializer{})
}
