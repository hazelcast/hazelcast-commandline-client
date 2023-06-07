package _multimap

import "github.com/hazelcast/hazelcast-commandline-client/internal/plug"

func GetTTL(ec plug.ExecContext) int64 {
	if _, ok := ec.Props().Get(multiMapTTL); ok {
		return ec.Props().GetInt(multiMapTTL)
	}
	return ttlUnset
}
