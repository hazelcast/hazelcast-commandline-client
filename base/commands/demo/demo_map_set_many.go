//go:build std || demo

package demo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type MapSetManyCmd struct{}

const (
	flagName           = "name"
	flagSize           = "size"
	argEntryCount      = "entryCount"
	argTitleEntryCount = "entry count"
	kb                 = "KB"
	mb                 = "MB"
	kbs                = 1024
	mbs                = kbs * 1024
)

func (m MapSetManyCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("map-setmany")
	help := "Generates multiple map entries"
	cc.SetCommandHelp(help, help)
	cc.AddStringFlag(flagName, "n", "default", false, "Name of the map.")
	cc.AddStringFlag(flagSize, "", "1", false, `Size of the map value in bytes, the following suffixes can also be used: kb, mb, e.g., 42kb)`)
	cc.AddInt64Arg(argEntryCount, argTitleEntryCount)
	return nil
}

func (m MapSetManyCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	count := ec.GetInt64Arg(argEntryCount)
	mapName := ec.Props().GetString(flagName)
	size := ec.Props().GetString(flagSize)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Creating entries in map %s with %d entries", mapName, count))
		mm, err := ci.Client().GetMap(ctx, mapName)
		if err != nil {
			return nil, err
		}
		return nil, createEntries(ctx, count, size, mm)
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Generated %d entries", count)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func createEntries(ctx context.Context, entryCount int64, size string, m *hazelcast.Map) error {
	v, err := makeValue(size)
	if err != nil {
		return err
	}
	for i := int64(1); i <= entryCount; i++ {
		k := fmt.Sprintf("k%d", i)
		err := m.Set(ctx, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeValue(size string) (string, error) {
	b, err := getValueSize(size)
	if err != nil {
		return "", err
	}
	return strings.Repeat("a", int(b)), nil
}

func getValueSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(sizeStr)
	if strings.HasSuffix(sizeStr, kb) {
		sizeStr = strings.TrimSuffix(sizeStr, kb)
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return 0, err
		}
		return size * kbs, nil
	}
	if strings.HasSuffix(sizeStr, mb) {
		sizeStr = strings.TrimSuffix(sizeStr, mb)
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return 0, err
		}
		return size * mbs, nil
	}
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func init() {
	Must(plug.Registry.RegisterCommand("demo:map-setmany", &MapSetManyCmd{}))
}
