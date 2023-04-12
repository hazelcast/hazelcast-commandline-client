//go:build base

package commands

import (
	"context"
	"fmt"
	"math"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type HomeCommand struct{}

func (hc HomeCommand) Init(cc plug.InitContext) error {
	short := "Print the CLC home directory"
	long := `Print the CLC home directory
	
If given, the arguments are joined as sub-paths.
	
Example:
	$ clc home foo bar
	/home/user/.hazelcast/foo/bar
`
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, math.MaxInt)
	cc.SetCommandUsage("home [subpath ...] [flags]")
	return nil
}

func (hc HomeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	dir := paths.Home()
	args := ec.Args()
	if len(args) > 0 {
		dir = filepath.Join(append([]string{dir}, args...)...)
	}
	I2(fmt.Fprintln(ec.Stdout(), dir))
	return nil
}

func (HomeCommand) Unwrappable() {}

func init() {
	Must(plug.Registry.RegisterCommand("home", &HomeCommand{}))
}
