//go:build base || list

package _list

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	listFlagName     = "name"
	listFlagShowType = "show-type"
	listPropertyName = "list"
)

type ListCommand struct {
}

func (mc *ListCommand) Init(cc plug.InitContext) error {
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.AddStringFlag(listFlagName, "n", defaultListName, false, "list name")
	cc.AddBoolFlag(listFlagShowType, "", false, false, "add the type names to the output")
	if !cc.Interactive() {
		cc.AddStringFlag(clc.PropertySchemaDir, "", paths.Schemas(), false, "set the schema directory")
	}
	cc.SetTopLevel(true)
	cc.SetCommandUsage("list [command] [flags]")
	help := "List operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *ListCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (mc *ListCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	props.SetBlocking(listPropertyName, func() (any, error) {
		listName := ec.Props().GetString(listFlagName)
		// empty list name is allowed
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		mv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			sp.SetText(fmt.Sprintf("Getting list %s", listName))
			m, err := ci.Client().GetList(ctx, listName)
			if err != nil {
				return nil, err
			}
			return m, nil
		})
		if err != nil {
			return nil, err
		}
		stop()
		return mv.(*hazelcast.List), nil
	})
	return nil
}

func init() {
	cmd := &ListCommand{}
	Must(plug.Registry.RegisterCommand("list", cmd))
	plug.Registry.RegisterAugmentor("20-list", cmd)
}
