//go:build base || config

package config

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type AddCmd struct{}

func (cm AddCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("add [configuration-name] [flags]")
	short := "Adds a configuration"
	long := `Adds a configuration with the given name/path and KEY=VALUE pairs
	
Overrides the previous configuration if it exists.
	
The following keys are supported:
	
	* cluster.name             STRING
	* cluster.address          STRING
	* cluster.user             STRING
	* cluster.password         STRING
	* cluster.discovery-token  STRING
	* ssl.enabled              BOOLEAN (true / false)
	* ssl.server               STRING
	* ssl.skip-verify          BOOLEAN (true / false)
	* ssl.ca-path              STRING
	* ssl.key-path             STRING
	* ssl.key-password         STRING
	* log.path                 STRING
	* log.level                ENUM (error / warn / info / debug)
	
`
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, math.MaxInt)
	return nil
}

func (cm AddCmd) Exec(_ context.Context, ec plug.ExecContext) error {
	target := ec.Args()[0]
	var opts clc.KeyValues[string, string]
	for _, arg := range ec.Args()[1:] {
		ps := strings.SplitN(arg, "=", 2)
		if len(ps) != 2 {
			return fmt.Errorf("invalid key=value pair: %s", arg)
		}
		opts = append(opts, clc.KeyValue[string, string]{
			Key:   ps[0],
			Value: ps[1],
		})
	}
	dir, cfgPath, err := config.Create(target, opts)
	if err != nil {
		return err
	}
	if ec.Interactive() || ec.Props().GetBool(clc.PropertyVerbose) {
		I2(fmt.Fprintf(ec.Stdout(), "Created configuration at: %s\n", filepath.Join(dir, cfgPath)))
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("config:add", &AddCmd{}))
}
