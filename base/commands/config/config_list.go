//go:build base

package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListCmd struct{}

func (cm ListCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list")
	long := fmt.Sprintf(`Lists known configurations
	
A known configuration is a directory at %s that contains config.yaml.
Directory names which start with . or _ are ignored.
`, paths.Configs())
	short := "Lists known configurations"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (cm ListCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	cd := paths.Configs()
	es, err := os.ReadDir(cd)
	if err != nil {
		return err
	}
	for _, e := range es {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		if paths.Exists(paths.Join(cd, e.Name(), "config.yaml")) {
			ec.AddOutputRows(output.Row{output.Column{
				Name:  "Config Name",
				Type:  serialization.TypeString,
				Value: e.Name(),
			}})
		}
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("config:list", &ListCmd{}))
}
