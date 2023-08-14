//go:build std || script

package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	prefixFile  = "file://"
	prefixHTTP  = "http://"
	prefixHTTPS = "https://"
)

type ScriptCommand struct{}

func (cm ScriptCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("script [path] [flags]")
	long := `Runs the script in the given local or HTTP location.
	
The script can contain:
	1. SQL statements
	2. CLC commands prefixed with backslash.
	3. Comments starting with -- (double dash)

The script should have either .clc or .sql extension.
Files with one of these two extensions are interpreted equivalently.
	
See examples/sql/dessert.sql for a sample script.
`
	short := "Runs the given script"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, 1)
	cc.AddBoolFlag(flagIgnoreErrors, "", false, false, "ignore errors during script execution")
	cc.AddBoolFlag(flagEcho, "", false, false, "print the executed command")
	return nil
}

func (cm ScriptCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	args := ec.Args()
	in := ec.Stdin()
	if len(args) > 0 {
		f, err := openScript(args[0])
		if err != nil {
			return fmt.Errorf("opening script: %w", err)
		}
		defer f.Close()
		in = f
	}
	sio := clc.IO{
		Stdin:  in,
		Stderr: ec.Stderr(),
		Stdout: ec.Stdout(),
	}
	m, err := ec.(*cmd.ExecContext).Main().Clone(false)
	if err != nil {
		return fmt.Errorf("cloning Main: %w", err)
	}
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	ie := ec.Props().GetBool(flagIgnoreErrors)
	echo := ec.Props().GetBool(flagEcho)
	textFn := makeTextFunc(m, ec, verbose, func(shortcut string) bool {
		// shortcuts are not supported in the script mode
		return false
	})
	sh := shell.NewOneshotShell(makeEndLineFunc(), sio, textFn)
	sh.SetIgnoreErrors(ie)
	sh.SetEcho(echo)
	sh.SetCommentPrefix("--")
	return sh.Run(ctx)
}

func (cm ScriptCommand) Unwrappable() {}

func openScript(location string) (io.ReadCloser, error) {
	if filepath.Ext(location) != ".clc" && filepath.Ext(location) != ".sql" {
		return nil, errors.New("the script should have either .clc or .sql extension")
	}
	if strings.HasPrefix(location, prefixFile) {
		location = location[len(prefixFile):]
		return os.Open(location)
	}
	if strings.HasPrefix(location, prefixHTTP) || strings.HasPrefix(location, prefixHTTPS) {
		resp, err := http.Get(location)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			return nil, errors.New(string(b))
		}
		return resp.Body, nil
	}
	return os.Open(location)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("script", &ScriptCommand{}))
}
