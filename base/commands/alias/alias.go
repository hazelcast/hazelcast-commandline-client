//go:build std || alias

package alias

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

var Aliases sync.Map

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
	Must(plug.Registry.RegisterCommand("alias", AliasCommand{}, plug.OnlyInteractive{}))
	Must(persistentAliases())
}

func persistentAliases() error {
	p := filepath.Join(paths.Home(), AliasFileName)
	data, err := os.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		if l == "" {
			continue
		}
		parts := strings.SplitN(l, "=", 2)
		Aliases.Store(parts[0], parts[1])
	}
	return nil
}
