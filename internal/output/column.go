package output

import (
	iserialization "github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type Column = iserialization.Column

func NewKeyColumn(kt int32, key any) Column {
	return Column{
		Name:  NameKey,
		Type:  kt,
		Value: key,
	}
}

func NewKeyTypeColumn(kt int32) Column {
	return Column{
		Name:  NameKeyType,
		Type:  iserialization.TypeString,
		Value: iserialization.TypeToLabel(kt),
	}
}

func NewValueColumn(vt int32, value any) Column {
	return Column{
		Name:  NameValue,
		Type:  vt,
		Value: value,
	}
}

func NewValueTypeColumn(vt int32) Column {
	return Column{
		Name:  NameValueType,
		Type:  iserialization.TypeString,
		Value: iserialization.TypeToLabel(vt),
	}
}
