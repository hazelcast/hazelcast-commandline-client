//go:build std || alias

package alias

import (
	"context"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const AliasFileName = "shell.clc"

type AliasCommand struct {
}

func (a AliasCommand) Init(cc plug.InitContext) error {
	cc.SetTopLevel(true)
	cc.SetCommandUsage("alias [command] [flags]")
	short := "Alias Operations"
	long := ` Users can defined aliases can be used in interactive mode and scripts.
Aliases can be used with '@' prefix. 
(eg. Assume user created an alias named "myAlias", which can be used as "@myAlias")
`
	cc.SetCommandHelp(long, short)
	return nil
}

func (a AliasCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("alias", AliasCommand{}))
}
