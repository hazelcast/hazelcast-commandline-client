package _map

import (
	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/make"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func MakeKeyData(ec plug.ExecContext, ci *hazelcast.ClientInternal) (hazelcast.Data, error) {
	ks := ec.Args()[0]
	kt := ec.Props().GetString(mapFlagKeyType)
	if kt == "" {
		kt = "string"
	}
	key, err := make.ValueFromString(ks, kt)
	if err != nil {
		return nil, err
	}
	return ci.EncodeData(key)
}

func MakeValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal) (hazelcast.Data, error) {
	vs := ec.Args()[1]
	vt := ec.Props().GetString(mapFlagValueType)
	if vt == "" {
		vt = "string"
	}
	value, err := make.ValueFromString(vs, vt)
	if err != nil {
		return nil, err
	}
	return ci.EncodeData(value)
}

func MakeKeyValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal) (hazelcast.Data, hazelcast.Data, error) {
	kd, err := MakeKeyData(ec, ci)
	if err != nil {
		return nil, nil, err
	}
	vd, err := MakeValueData(ec, ci)
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
