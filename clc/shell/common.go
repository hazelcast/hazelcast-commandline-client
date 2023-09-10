package shell

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/sql"

	"github.com/hazelcast/hazelcast-commandline-client/base/maps"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	clcsql "github.com/hazelcast/hazelcast-commandline-client/clc/sql"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const CmdPrefix = `\`

var ErrHelp = errors.New("interactive help")

func ConvertStatement(ctx context.Context, ec plug.ExecContext, stmt string) (func() error, error) {
	var query string
	stmt = strings.TrimSpace(stmt)
	if strings.HasPrefix(stmt, "help") {
		return nil, ErrHelp
	}
	if strings.HasPrefix(stmt, CmdPrefix) {
		// this is a shell command
		stmt = strings.TrimPrefix(stmt, CmdPrefix)
		parts := strings.Fields(stmt)
		switch parts[0] {
		case "di":
			if len(parts) == 1 {
				return func() error {
					return maps.Indexes(ctx, ec, "")
				}, nil
			}
			if len(parts) == 2 {
				return func() error {
					return maps.Indexes(ctx, ec, parts[1])
				}, nil
			} else {
				return nil, fmt.Errorf("Usage: %sdi [mapping]", CmdPrefix)
			}
		case "dm":
			if len(parts) == 1 {
				query = "show mappings;"
			} else if len(parts) == 2 {
				// escape single quote
				mn := strings.Replace(parts[1], "'", "''", -1)
				query = fmt.Sprintf(`
					SELECT * FROM information_schema.mappings
					WHERE table_name = '%s';
				`, mn)
			} else {
				return nil, fmt.Errorf("Usage: %sdm [mapping]", CmdPrefix)
			}
		case "dm+":
			if len(parts) == 1 {
				query = "show mappings;"
			} else if len(parts) == 2 {
				// escape single quote
				mn := strings.Replace(parts[1], "'", "''", -1)
				query = fmt.Sprintf(`
					SELECT * FROM information_schema.columns
					WHERE table_name = '%s';
				`, mn)
			} else {
				return nil, fmt.Errorf("Usage: %sdm+ [mapping]", CmdPrefix)
			}
		case "exit":
			return nil, ErrExit
		default:
			return nil, fmt.Errorf("Unknown shell command: %s", stmt)
		}
	} else {
		query = stmt
	}
	f := func() error {
		resV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			res, err := clcsql.ExecSQL(ctx, ec, sp, query)
			if err != nil {
				return nil, err
			}
			return res, nil
		})
		if err != nil {
			return err
		}
		defer stop()
		res := resV.(sql.Result)
		if err := clcsql.UpdateOutput(ctx, ec, res); err != nil {
			return err
		}
		return nil
	}
	return f, nil
}

func InteractiveHelp() string {
	return `
Shortcut Commands:
	\di           List Indexes
	\di  MAPPING  List Indexes for a specific mapping
	\dm           List mappings
	\dm  MAPPING  Display information about a mapping
	\dm+ MAPPING  Describe a mapping
	\exit         Exit the shell
	\help         Display help for CLC commands
`
}
