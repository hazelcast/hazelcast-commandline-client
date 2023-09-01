package alias

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
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
		Aliases.Delete(name)
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("alias:remove", &AliasRemoveCommand{}, plug.OnlyInteractive{}))
}
