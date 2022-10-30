package ingest

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/serialization"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	mapName = "name"
	index   = "index"
)

type IngestCommand struct{}

func (cm IngestCommand) Init(cc plug.InitContext) error {
	cc.AddCommandGroup("batch", "Batch Operations")
	cc.SetCommandGroup("batch")
	cc.SetCommandUsage("ingest FILE")
	help := "read from the given file and write to the given map"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(mapName, "n", "", true, "map to update")
	cc.AddIntFlag(index, "", 1, false, "set the index start")
	return nil
}

func (cm IngestCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	// TODO: proper error handling
	path := ec.Args()[0]
	if !strings.HasSuffix(path, ".json") {
		return fmt.Errorf("only files having the .json extension are supported")
	}
	format := ec.Props().GetString(clc.PropertyFormat)
	if format != "json" {
		return fmt.Errorf("only JSON data line per record format is supported")
	}
	ci := MustValue(ec.ClientInternal(ctx))
	m := MustValue(ci.Client().GetMap(ctx, ec.Props().GetString(mapName)))
	f := MustValue(os.Open(path))
	start := ec.Props().GetInt(index) - 1
	idx := start
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		idx++
		text := sc.Text()
		// TODO: check the line contains valid JSON
		var v any
		if err := json.Unmarshal([]byte(text), &v); err != nil {
			fmt.Fprintf(ec.Stderr(), "invalid JSON in file: %s at line: %d\n", path, idx-start)
			continue
		}
		if err := m.Set(ctx, idx, serialization.JSON(text)); err != nil {
			ec.Logger().Error(err)
		}
	}
	// return value and error is ignored
	fmt.Fprintf(ec.Stdout(), "Ingested %d records\n", idx-start)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("ingest", &IngestCommand{}))
}
