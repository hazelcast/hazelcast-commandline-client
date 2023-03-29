package base

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type GlobalInitializer struct{}

func (g GlobalInitializer) Init(cc plug.InitContext) error {
	// base group IDs
	cc.AddCommandGroup(clc.GroupDDSID, "Distributed Data Structures")
	// output type flag
	pns := plug.Registry.PrinterNames()
	slices.Sort(pns)
	usage := fmt.Sprintf("set the output type, one of: %s", strings.Join(pns, ", "))
	// format is delimited for command line mode.
	var format string
	if slices.Contains(pns, PrinterDelimited) {
		format = PrinterDelimited
	}
	// format is table for the interactive mode.
	if cc.Interactive() && !shell.IsPipe() {
		if slices.Contains(pns, PrinterTable) {
			format = PrinterTable
		}
	}
	// other flags
	cc.AddStringFlag(clc.PropertyFormat, clc.ShortcutFormat, format, false, usage)
	cc.AddBoolFlag(clc.PropertyVerbose, "", false, false, "enable verbose output")
	cc.AddBoolFlag(clc.PropertyQuite, "", false, false, "disable unnecessary output")
	lp := paths.DefaultLogPath(time.Now())
	if !cc.Interactive() {
		cc.AddStringFlag(clc.PropertyConfig, clc.ShortcutConfig, "", false, "set the configuration")
		cc.AddStringFlag(clc.PropertyLogPath, "", lp, false, "set the log path, use stderr to log to stderr")
		cc.AddStringFlag(clc.PropertyLogLevel, "", "info", false, "set the log level")
	}
	// configuration
	cc.AddStringConfig(clc.PropertyClusterAddress, "localhost:5701", "", "cluster address")
	cc.AddStringConfig(clc.PropertyClusterName, "dev", "", "cluster name")
	cc.AddStringConfig(clc.PropertyLogPath, "", clc.PropertyLogPath, "log path")
	cc.AddStringConfig(clc.PropertyLogLevel, "", clc.PropertyLogLevel, "log level")
	cc.AddStringConfig(clc.PropertySchemaDir, "", clc.PropertySchemaDir, "schema directory")
	cc.AddStringConfig(clc.PropertyClusterDiscoveryToken, "", "", "Viridian token")
	return nil
}

func init() {
	plug.Registry.RegisterGlobalInitializer("00-global-init", &GlobalInitializer{})
}
