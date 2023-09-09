//go:build std || set

package set

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

func addValueTypeFlag(cc plug.InitContext) {
	help := fmt.Sprintf("value type (one of: %s)", strings.Join(internal.SupportedTypeNames, ", "))
	cc.AddStringFlag(setFlagValueType, "v", "string", false, help)
}

func makeValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal, valueStr string) (hazelcast.Data, error) {
	vt := ec.Props().GetString(setFlagValueType)
	if vt == "" {
		vt = "string"
	}
	value, err := mk.ValueFromString(valueStr, vt)
	if err != nil {
		return nil, err
	}
	return ci.EncodeData(value)
}

func stringToPartitionID(ci *hazelcast.ClientInternal, name string) (int32, error) {
	var partitionID int32
	var keyData hazelcast.Data
	var err error
	idx := strings.Index(name, "@")
	if keyData, err = ci.EncodeData(name[idx+1:]); err != nil {
		return 0, err
	}
	if partitionID, err = ci.GetPartitionID(keyData); err != nil {
		return 0, err
	}
	return partitionID, nil
}

func getSet(ctx context.Context, ec plug.ExecContext, sp clc.Spinner) (*hazelcast.Set, error) {
	name := ec.Props().GetString(base.FlagName)
	ci, err := cmd.ClientInternal(ctx, ec, sp)
	if err != nil {
		return nil, err
	}
	sp.SetText(fmt.Sprintf("Getting Set '%s'", name))
	return ci.Client().GetSet(ctx, name)
}
