//go:build std || alias

package alias

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
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
		return nil, CreateAlias(name, cmd)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func CreateAlias(key, value string) error {
	p := filepath.Join(paths.Home(), AliasFileName)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			_, err = os.Create(p)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	lines := strings.Split(string(data), "\n")
	var newData []string
	updated := false
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			if parts[0] == key {
				newData = append(newData, fmt.Sprintf("%s=%s", key, value))
				updated = true
			} else {
				newData = append(newData, line)
			}
		}
	}
	if !updated {
		newData = append(newData, fmt.Sprintf("%s=%s", key, value))
	}
	c := strings.Join(newData, "\n")
	return os.WriteFile(filepath.Join(paths.Home(), AliasFileName), []byte(c), 0644)
}

func init() {
	Must(plug.Registry.RegisterCommand("alias:add", &AliasAddCommand{}))
}
