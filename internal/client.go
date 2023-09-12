package internal

import (
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
)

func StringToPartitionID(ci *hazelcast.ClientInternal, name string) (int32, error) {
	idx := strings.Index(name, "@")
	keyData, err := ci.EncodeData(name[idx+1:])
	if err != nil {
		return 0, err
	}
	partitionID, err := ci.GetPartitionID(keyData)
	if err != nil {
		return 0, err
	}
	return partitionID, nil
}
