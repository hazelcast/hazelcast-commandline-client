package alias

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
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
		return listAliases()
	})
	if err != nil {
		return err
	}
	stop()
	aliases := sv.([]alias)
	if len(aliases) == 0 {
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

func listAliases() ([]alias, error) {
	var aliases []alias
	data, err := os.ReadFile(filepath.Join(paths.Home(), AliasFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		if l == "" {
			continue
		}
		parts := strings.SplitN(l, "=", 2)
		aliases = append(aliases, alias{
			name:  parts[0],
			value: parts[1],
		})
	}
	return aliases, nil
}

func init() {
	Must(plug.Registry.RegisterCommand("alias:list", &AliasListCommand{}))
}
