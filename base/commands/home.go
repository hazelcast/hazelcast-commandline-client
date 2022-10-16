package commands

import (
	"fmt"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type HomeCommand struct{}

func (hc HomeCommand) Init(cc plug.InitContext) error {
	usage := "Print the CLC home directory"
	cc.SetCommandUsage(usage, usage)
	cc.SetPositionalArgCount(0, 1)
	return nil
}

func (hc HomeCommand) Exec(ec plug.ExecContext) error {
	dir := paths.HomeDir()
	args := ec.Args()
	if len(args) > 0 {
		dir = filepath.Join(append([]string{dir}, args...)...)
	}
	I2(fmt.Fprintln(ec.Stdout(), dir))
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("home", &HomeCommand{}))
}
