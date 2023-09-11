//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct{}

func (Command) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("viridian")
	cc.AddCommandGroup("viridian", "Viridian")
	cc.SetCommandGroup("viridian")
	cc.SetTopLevel(true)
	help := "Various Viridian operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (Command) Exec(ctx context.Context, ec plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("viridian", &Command{}))
}
