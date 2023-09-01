//go:build std || alias

package alias

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type AliasAddCommand struct{}

func (a AliasAddCommand) Init(cc plug.InitContext) error {
	cc.SetPositionalArgCount(2, math.MaxInt)
	short := "Add an alias with the given name and command"
	long := ` Add an alias with the given name and command.
The command must be given in double quotes (eg. alias add myAlias "map set 1 2").
`
	cc.SetCommandHelp(long, short)
	cc.SetCommandUsage("add [name] [command]")
	return nil
}

func (a AliasAddCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Args()[0]
	cmd := strings.Join(ec.Args()[1:], " ")
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Saving alias %s", name))
		Aliases.Store(name, cmd)
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("alias:add", &AliasAddCommand{}, plug.OnlyInteractive{}))
}
