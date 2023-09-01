package alias

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type AliasListCommand struct{}

func (a AliasListCommand) Init(cc plug.InitContext) error {
	cc.SetPositionalArgCount(0, 0)
	help := "list user defined aliases"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("list")
	return nil
}

type alias struct {
	name  string
	value string
}

func (a AliasListCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Listing aliases")
		var all []alias
		Aliases.Range(func(key, value any) bool {
			all = append(all, alias{
				name:  key.(string),
				value: value.(string),
			})
			return true
		})
		return all, nil
	})
	if err != nil {
		return err
	}
	stop()
	all := sv.([]alias)
	if len(all) == 0 {
		ec.PrintlnUnnecessary("No aliases found.")
		return nil
	}
	var rows []output.Row
	for _, a := range sv.([]alias) {
		rows = append(rows, output.Row{
			output.Column{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: a.name,
			},
			output.Column{
				Name:  "Value",
				Type:  serialization.TypeString,
				Value: a.value,
			},
		})
	}
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	Must(plug.Registry.RegisterCommand("alias:list", &AliasListCommand{}, plug.OnlyInteractive{}))
}
