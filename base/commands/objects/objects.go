package objects

import (
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ObjectsCommand struct{}

func (cm ObjectsCommand) Init(cc plug.InitContext) error {
	cc.SetCommandGroup("dds")
	cc.SetTopLevel(true)
	help := "Generic distributed data structure operations"
	cc.SetCommandUsage("objects [command]")
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm ObjectsCommand) Exec(ec plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("objects", &ObjectsCommand{}))
}
