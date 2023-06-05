//go:build base || multimap

package _multimap

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
)

const (
	multiMapFlagName     = "name"
	multiMapFlagShowType = "show-type"
	multiMapPropertyName = "multiMap"
)

type MultiMapCommand struct {
}

func (m MultiMapCommand) Init(cc plug.InitContext) error {
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.AddStringFlag(multiMapFlagName, "n", defaultMultiMapName, false, "multiMap name")
	cc.AddBoolFlag(multiMapFlagShowType, "", false, false, "add the type names to the output")
	if !cc.Interactive() {
		cc.AddStringFlag(clc.PropertySchemaDir, "", paths.Schemas(), false, "set the schema directory")
	}
	cc.SetTopLevel(true)
	cc.SetCommandUsage("multiMap [command] [flags]")
	help := "MultiMap operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (m MultiMapCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (m MultiMapCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	props.SetBlocking(multiMapPropertyName, func() (any, error) {
		mmName := ec.Props().GetString(multiMapFlagName)
		// empty multiMap name is allowed
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		mv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			sp.SetText(fmt.Sprintf("Getting multiMap %s", mmName))
			m, err := ci.Client().GetMap(ctx, mmName)
			if err != nil {
				return nil, err
			}
			return m, nil
		})
		if err != nil {
			return nil, err
		}
		stop()
		return mv.(*hazelcast.MultiMap), nil
	})
	return nil
}

func init() {
	cmd := &MultiMapCommand{}
	Must(plug.Registry.RegisterCommand("multiMap", cmd))
	plug.Registry.RegisterAugmentor("20-multiMap", cmd)
}
