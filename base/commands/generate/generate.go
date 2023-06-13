package generate

import (
	"context"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type GenerateCommand struct{}

func (gc *GenerateCommand) Init(cc plug.InitContext) error {
	cc.SetTopLevel(true)
	cc.SetCommandUsage("generate [command] [flags]")
	help := "Generate commands"
	cc.SetCommandHelp(help, help)
	return nil
}

func (gc GenerateCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("generate", &GenerateCommand{}))
}
