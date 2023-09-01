package alias

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type AliasRemoveCommand struct{}

func (a AliasRemoveCommand) Init(cc plug.InitContext) error {
	cc.SetPositionalArgCount(1, 1)
	help := "remove an alias with the given name"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("remove [name] [command]")
	return nil
}

func (a AliasRemoveCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Args()[0]
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Removing alias %s", name))
		return nil, removeAlias(name)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func removeAlias(name string) error {
	data, err := os.ReadFile(filepath.Join(paths.Home(), AliasFileName))
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	var newData []string
	for _, l := range lines {
		parts := strings.SplitN(l, "=", 2)
		if parts[0] != name {
			newData = append(newData, l)
		}
	}
	newContent := strings.Join(newData, "\n")
	return os.WriteFile(filepath.Join(paths.Home(), AliasFileName), []byte(newContent), 0600)
}

func init() {
	Must(plug.Registry.RegisterCommand("alias:remove", &AliasRemoveCommand{}))
}
