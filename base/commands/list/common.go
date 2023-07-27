//go:build std || list

package list

import (
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func addValueTypeFlag(cc plug.InitContext) {
	help := fmt.Sprintf("value type (one of: %s)", strings.Join(internal.SupportedTypeNames, ", "))
	cc.AddStringFlag(listFlagValueType, "v", "string", false, help)
}

func makeValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal, valueStr string) (hazelcast.Data, error) {
	vt := ec.Props().GetString(listFlagValueType)
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
