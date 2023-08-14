//go:build std || map

package _map

import (
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func GetTTL(ec plug.ExecContext) int64 {
	if _, ok := ec.Props().Get(mapTTL); ok {
		return ec.Props().GetInt(mapTTL)
	}
	return ttlUnset
}

func GetMaxIdle(ec plug.ExecContext) int64 {
	if _, ok := ec.Props().Get(mapMaxIdle); ok {
		return ec.Props().GetInt(mapMaxIdle)
	}
	return ttlUnset
}
