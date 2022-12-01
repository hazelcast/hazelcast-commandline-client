package output

import iserialization "github.com/hazelcast/hazelcast-commandline-client/internal/serialization"

func convertColumn(col Column) any {
	switch col.Type {
	case iserialization.TypeByte, iserialization.TypeBool, iserialization.TypeUInt16,
		iserialization.TypeInt16, iserialization.TypeInt32, iserialization.TypeInt64,
		iserialization.TypeFloat32, iserialization.TypeFloat64, iserialization.TypeString:
		return col.Value
	case iserialization.TypeNil:
		return ValueNil
	case iserialization.TypeUnknown:
		return ValueUnknown
	case iserialization.TypeSkip:
		return ValueSkip
	default:
		return col.SingleLine()
	}
}
