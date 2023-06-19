package demo

import (
	"context"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Cmd struct{}

func (cm *Cmd) Init(cc plug.InitContext) error {
	cc.AddCommandGroup(clc.GroupDemoID, "Demo")
	cc.SetCommandGroup(clc.GroupDemoID)
	cc.SetTopLevel(true)
	help := "Demo operations"
	cc.SetCommandUsage("demo [command]")
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm *Cmd) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (cm *Cmd) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	props.SetBlocking(demoMapPropertyName, func() (any, error) {
		keyVals, err := keyValMap(ec)
		if err != nil {
			return nil, err
		}
		mapName, ok := keyVals[pairMapName]
		if !ok {
			return nil, fmt.Errorf("%s key-value pair must be given", pairMapName)
		}
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		mv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			sp.SetText(fmt.Sprintf("Getting map %s", mapName))
			m, err := ci.Client().GetMap(ctx, mapName)
			if err != nil {
				return nil, err
			}
			return m, nil
		})
		if err != nil {
			return nil, err
		}
		stop()
		return mv.(*hazelcast.Map), nil
	})
	return nil
}

func keyValMap(ec plug.ExecContext) (map[string]string, error) {
	keyVals := map[string]string{}
	for _, keyval := range ec.Args()[1:] {
		kv := strings.Split(keyval, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("Key-value pair is incorrect %s", keyval)
		}
		keyVals[kv[0]] = kv[1]
	}
	return keyVals, nil
}

func init() {
	cmd := &Cmd{}
	Must(plug.Registry.RegisterCommand("demo", cmd))
	plug.Registry.RegisterAugmentor("20-demo", cmd)

}
