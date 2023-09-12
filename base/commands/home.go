//go:build std || home

package commands

import (
	"context"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const (
	argSubPath      = "subpath"
	argTitleSubPath = "subpath"
)

type HomeCommand struct{}

func (HomeCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("home")
	short := "Print the CLC home directory"
	long := `Print the CLC home directory
	
If given, the arguments are joined as sub-paths.
	
Example:
	$ clc home foo bar
	/home/user/.hazelcast/foo/bar
`
	cc.SetCommandHelp(long, short)
	cc.AddStringSliceArg(argSubPath, argTitleSubPath, 0, clc.MaxArgs)
	return nil
}

func (HomeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	path := paths.Home()
	args := ec.GetStringSliceArg(argSubPath)
	if len(args) > 0 {
		path = filepath.Join(append([]string{path}, args...)...)
	}
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "Path",
			Type:  serialization.TypeString,
			Value: path,
		},
	})
}

func init() {
	check.Must(plug.Registry.RegisterCommand("home", &HomeCommand{}))
}
