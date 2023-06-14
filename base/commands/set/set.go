//go:build base || set

package set

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
)

const (
	setFlagName     = "name"
	setFlagShowType = "show-type"
	setPropertyName = "set"
)

type SetCommand struct{}

func (sc *SetCommand) Init(cc plug.InitContext) error {
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.AddStringFlag(setFlagName, "n", defaultSetName, false, "set name")
	cc.AddBoolFlag(setFlagShowType, "", false, false, "add the type names to the output")
	if !cc.Interactive() {
		cc.AddStringFlag(clc.PropertySchemaDir, "", paths.Schemas(), false, "set the schema directory")
	}
	cc.SetTopLevel(true)
	cc.SetCommandUsage("set [command] [flags]")
	help := "Set operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (sc *SetCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (sc *SetCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	props.SetBlocking(setPropertyName, func() (any, error) {
		setName := ec.Props().GetString(setFlagName)
		// empty set name is allowed
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		val, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			sp.SetText(fmt.Sprintf("Getting set %s", setName))
			q, err := ci.Client().GetSet(ctx, setName)
			if err != nil {
				return nil, err
			}
			return q, nil
		})
		if err != nil {
			return nil, err
		}
		stop()
		return val.(*hazelcast.Set), nil
	})
	return nil
}

func init() {
	cmd := &SetCommand{}
	check.Must(plug.Registry.RegisterCommand("set", cmd))
	plug.Registry.RegisterAugmentor("20-set", cmd)
}
