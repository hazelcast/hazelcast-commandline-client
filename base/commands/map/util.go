//go:build base || map

package _map

import (
	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func MakeKeyData(ec plug.ExecContext, ci *hazelcast.ClientInternal, keyStr string) (hazelcast.Data, error) {
	kt := ec.Props().GetString(mapFlagKeyType)
	if kt == "" {
		kt = "string"
	}
	key, err := mk.ValueFromString(keyStr, kt)
	if err != nil {
		return nil, err
	}
	return ci.EncodeData(key)
}

func MakeValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal, valueStr string) (hazelcast.Data, error) {
	vt := ec.Props().GetString(mapFlagValueType)
	if vt == "" {
		vt = "string"
	}
	value, err := mk.ValueFromString(valueStr, vt)
	if err != nil {
		return nil, err
	}
	return ci.EncodeData(value)
}

func MakeKeyValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal, keyStr, valueStr string) (hazelcast.Data, hazelcast.Data, error) {
	kd, err := MakeKeyData(ec, ci, keyStr)
	if err != nil {
		return nil, nil, err
	}
	vd, err := MakeValueData(ec, ci, valueStr)
	if err != nil {
		return nil, nil, err
	}
	return kd, vd, nil
}

func GetTTL(ec plug.ExecContext) int64 {
	if _, ok := ec.Props().Get(mapTTL); ok {
		return ec.Props().GetInt(mapTTL)
	}
	return ttlUnset
}
