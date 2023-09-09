//go:build std || multimap

package multimap

import (
	"context"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func addKeyTypeFlag(cc plug.InitContext) {
	help := fmt.Sprintf("key type (one of: %s)", strings.Join(internal.SupportedTypeNames, ", "))
	cc.AddStringFlag(multiMapFlagKeyType, "k", "string", false, help)
}

func addValueTypeFlag(cc plug.InitContext) {
	help := fmt.Sprintf("value type (one of: %s)", strings.Join(internal.SupportedTypeNames, ", "))
	cc.AddStringFlag(multiMapFlagValueType, "v", "string", false, help)
}

func makeKeyData(ec plug.ExecContext, ci *hazelcast.ClientInternal, keyStr string) (hazelcast.Data, error) {
	kt := ec.Props().GetString(multiMapFlagKeyType)
	if kt == "" {
		kt = "string"
	}
	key, err := mk.ValueFromString(keyStr, kt)
	if err != nil {
		return nil, err
	}
	return ci.EncodeData(key)
}

func makeValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal, valueStr string) (hazelcast.Data, error) {
	vt := ec.Props().GetString(multiMapFlagValueType)
	if vt == "" {
		vt = "string"
	}
	value, err := mk.ValueFromString(valueStr, vt)
	if err != nil {
		return nil, err
	}
	return ci.EncodeData(value)
}

func makeKeyValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal, keyStr, valueStr string) (hazelcast.Data, hazelcast.Data, error) {
	kd, err := makeKeyData(ec, ci, keyStr)
	if err != nil {
		return nil, nil, err
	}
	vd, err := makeValueData(ec, ci, valueStr)
	if err != nil {
		return nil, nil, err
	}
	return kd, vd, nil
}

func getMultiMap(ctx context.Context, ec plug.ExecContext, sp clc.Spinner) (*hazelcast.MultiMap, error) {
	name := ec.Props().GetString(base.FlagName)
	ci, err := cmd.ClientInternal(ctx, ec, sp)
	if err != nil {
		return nil, err
	}
	sp.SetText(fmt.Sprintf("Getting MultiMap '%s'", name))
	return ci.Client().GetMultiMap(ctx, name)
}
