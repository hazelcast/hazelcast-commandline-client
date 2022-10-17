package make

import "github.com/hazelcast/hazelcast-commandline-client/internal"

func ValueFromString(value, valueType string) (any, error) {
	return internal.ConvertString(value, valueType)
}
