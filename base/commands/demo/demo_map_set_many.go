//go:build std || demo

package demo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
)

type MapSetManyCmd struct{}

const (
	flagName = "name"
	flagSize = "size"
	kb       = "KB"
	mb       = "MB"
	kbs      = 1024
	mbs      = kbs * 1024
)

func (m MapSetManyCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("map-setmany [entry-count] [flags]")
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(flagName, "n", "default", false, "Name of the map.")
	cc.AddStringFlag(flagSize, "", "1", false, `Maybe an integer which is number of bytes, or one of the following:
Xkb: kilobytes
Xmb: megabytes
`)
	help := "Generates multiple map entries."
	cc.SetCommandHelp(help, help)
	return nil
}

func (m MapSetManyCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	entryCount := ec.Args()[0]
	c, err := strconv.Atoi(entryCount)
	if err != nil {
		return err
	}
	mapName := ec.Props().GetString(flagName)
	size := ec.Props().GetString(flagSize)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Creating entries in map %s with %d entries", mapName, c))
		mm, err := ci.Client().GetMap(ctx, mapName)
		if err != nil {
			return nil, err
		}
		return nil, createEntries(ctx, c, size, mm)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func createEntries(ctx context.Context, entryCount int, size string, m *hazelcast.Map) error {
	v, err := createVal(size)
	if err != nil {
		return err
	}
	for i := 1; i <= entryCount; i++ {
		k := fmt.Sprintf("k%d", i)
		err := m.Set(ctx, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func createVal(size string) (string, error) {
	b, err := byteSize(size)
	if err != nil {
		return "", err
	}
	return strings.Repeat("a", int(b)), nil
}

func byteSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
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
