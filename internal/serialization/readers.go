package serialization

import (
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
)

func boolToColumn(name string, v bool) Column {
	return Column{
		Name:  name,
		Type:  TypeBool,
		Value: v,
	}
}

func int8ToColumn(name string, v int8) Column {
	return Column{
		Name:  name,
		Type:  TypeInt8,
		Value: v,
	}
}

func int16ToColumn(name string, v int16) Column {
	return Column{
		Name:  name,
		Type:  TypeInt16,
		Value: v,
	}
}

func int32ToColumn(name string, v int32) Column {
	return Column{
		Name:  name,
		Type:  TypeInt32,
		Value: v,
	}
}

func int64ToColumn(name string, v int64) Column {
	return Column{
		Name:  name,
		Type:  TypeInt64,
		Value: v,
	}
}

func float32ToColumn(name string, v float32) Column {
	return Column{
		Name:  name,
		Type:  TypeFloat32,
		Value: v,
	}
}

func float64ToColumn(name string, v float64) Column {
	return Column{
		Name:  name,
		Type:  TypeFloat64,
		Value: v,
	}
}

func ptrStringToColumn(name string, v *string) Column {
	return Column{
		Name:  name,
		Type:  TypeString,
		Value: v,
	}
}

func ptrDecimalToColumn(name string, v *types.Decimal) Column {
	return Column{
		Name:  name,
		Type:  TypeJavaDecimal,
		Value: v,
	}
}

func ptrLocalTimeToColumn(name string, v *types.LocalTime) Column {
	return Column{
		Name:  name,
		Type:  TypeJavaLocalTime,
		Value: v,
	}
}

func ptrLocalDateToColumn(name string, v *types.LocalDate) Column {
	return Column{
		Name:  name,
		Type:  TypeJavaLocalDate,
		Value: v,
	}
}

func ptrLocalDateTimeToColumn(name string, v *types.LocalDateTime) Column {
	return Column{
		Name:  name,
		Type:  TypeJavaLocalDateTime,
		Value: v,
	}
}

func ptrOffsetDateTimeToColumn(name string, v *types.OffsetDateTime) Column {
	return Column{
		Name:  name,
		Type:  TypeJavaOffsetDateTime,
		Value: v,
	}
}

func compactToColumn(name string, v any) Column {
	return Column{
		Name:  name,
		Type:  TypeCompact,
		Value: v,
	}
}

func arrayToColumn[T any](name string, arrayType int32, vs []T, f func(name string, v T) Column) Column {
	cs := make(ColumnList, len(vs))
	for i, v := range vs {
		cs[i] = f("", v)
	}
	return Column{
		Name:  name,
		Type:  arrayType,
		Value: cs,
	}
}

func nullableToColumn[T any](name string, v *T, f func(name string, v T) Column) Column {
	if v == nil {
		return Column{
			Name: name,
			Type: TypeNil,
		}
	}
	return f(name, *v)
}

func nullableArrayToColumn[T any](name string, arrayType int32, vs []*T, f func(name string, v T) Column) Column {
	cs := make(ColumnList, len(vs))
	for i, v := range vs {
		cs[i] = nullableToColumn("", v, f)
	}
	return Column{
		Name:  name,
		Type:  arrayType,
		Value: cs,
	}
}

var compactReaders = map[serialization.FieldKind]compactFieldReader{
	serialization.FieldKindBoolean: func(r serialization.CompactReader, field string) Column {
		return boolToColumn(field, r.ReadBoolean(field))
	},
	serialization.FieldKindArrayOfBoolean: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfBoolean(field)
		return arrayToColumn(field, TypeBoolArray, vs, boolToColumn)
	},
	serialization.FieldKindInt8: func(r serialization.CompactReader, field string) Column {
		return int8ToColumn(field, r.ReadInt8(field))
	},
	serialization.FieldKindArrayOfInt8: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfInt8(field)
		return arrayToColumn(field, TypeInt8Array, vs, int8ToColumn)
	},
	// FieldKindChar        : Not decoded due to spec.
	// FieldKindArrayOfChar : Not decoded due to spec.
	serialization.FieldKindInt16: func(r serialization.CompactReader, field string) Column {
		return int16ToColumn(field, r.ReadInt16(field))
	},
	serialization.FieldKindArrayOfInt16: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfInt16(field)
		return arrayToColumn(field, TypeInt16Array, vs, int16ToColumn)
	},
	serialization.FieldKindInt32: func(r serialization.CompactReader, field string) Column {
		return int32ToColumn(field, r.ReadInt32(field))
	},
	serialization.FieldKindArrayOfInt32: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfInt32(field)
		return arrayToColumn(field, TypeInt32Array, vs, int32ToColumn)
	},
	serialization.FieldKindInt64: func(r serialization.CompactReader, field string) Column {
		return int64ToColumn(field, r.ReadInt64(field))
	},
	serialization.FieldKindArrayOfInt64: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfInt64(field)
		return arrayToColumn(field, TypeInt64Array, vs, int64ToColumn)
	},
	serialization.FieldKindFloat32: func(r serialization.CompactReader, field string) Column {
		return float32ToColumn(field, r.ReadFloat32(field))
	},
	serialization.FieldKindArrayOfFloat32: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfFloat32(field)
		return arrayToColumn(field, TypeFloat32Array, vs, float32ToColumn)
	},
	serialization.FieldKindFloat64: func(r serialization.CompactReader, field string) Column {
		return float64ToColumn(field, r.ReadFloat64(field))
	},
	serialization.FieldKindArrayOfFloat64: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfFloat64(field)
		return arrayToColumn(field, TypeFloat64Array, vs, float64ToColumn)
	},
	serialization.FieldKindString: func(r serialization.CompactReader, field string) Column {
		return ptrStringToColumn(field, r.ReadString(field))
	},
	serialization.FieldKindArrayOfString: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfString(field)
		return arrayToColumn(field, TypeStringArray, vs, ptrStringToColumn)
	},
	serialization.FieldKindDecimal: func(r serialization.CompactReader, field string) Column {
		return ptrDecimalToColumn(field, r.ReadDecimal(field))
	},
	serialization.FieldKindArrayOfDecimal: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfDecimal(field)
		return arrayToColumn(field, TypeDecimalArray, vs, ptrDecimalToColumn)
	},
	serialization.FieldKindTime: func(r serialization.CompactReader, field string) Column {
		return ptrLocalTimeToColumn(field, r.ReadTime(field))
	},
	serialization.FieldKindArrayOfTime: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfTime(field)
		return arrayToColumn(field, TypeJavaLocalTimeArray, vs, ptrLocalTimeToColumn)
	},
	serialization.FieldKindDate: func(r serialization.CompactReader, field string) Column {
		return ptrLocalDateToColumn(field, r.ReadDate(field))
	},
	serialization.FieldKindArrayOfDate: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfDate(field)
		return arrayToColumn(field, TypeJavaLocalDateArray, vs, ptrLocalDateToColumn)
	},
	serialization.FieldKindTimestamp: func(r serialization.CompactReader, field string) Column {
		return ptrLocalDateTimeToColumn(field, r.ReadTimestamp(field))
	},
	serialization.FieldKindArrayOfTimestamp: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfTimestamp(field)
		return arrayToColumn(field, TypeJavaLocalDateTimeArray, vs, ptrLocalDateTimeToColumn)
	},
	serialization.FieldKindTimestampWithTimezone: func(r serialization.CompactReader, field string) Column {
		return ptrOffsetDateTimeToColumn(field, r.ReadTimestampWithTimezone(field))
	},
	serialization.FieldKindArrayOfTimestampWithTimezone: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfTimestampWithTimezone(field)
		return arrayToColumn(field, TypeJavaOffsetDateTimeArray, vs, ptrOffsetDateTimeToColumn)
	},
	serialization.FieldKindCompact: func(r serialization.CompactReader, field string) Column {
		return compactToColumn(field, r.ReadCompact(field))
	},
	serialization.FieldKindArrayOfCompact: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfCompact(field)
		return arrayToColumn(field, TypeCompactArray, vs, compactToColumn)
	},
	// FieldKindPortable        : Not decoded due to spec.
	// FieldKindArrayOfPortable : Not decoded due to spec.
	serialization.FieldKindNullableBoolean: func(r serialization.CompactReader, field string) Column {
		v := r.ReadNullableBoolean(field)
		return nullableToColumn(field, v, boolToColumn)
	},
	serialization.FieldKindArrayOfNullableBoolean: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfNullableBoolean(field)
		return nullableArrayToColumn(field, TypeBoolArray, vs, boolToColumn)
	},
	serialization.FieldKindNullableInt8: func(r serialization.CompactReader, field string) Column {
		v := r.ReadNullableInt8(field)
		return nullableToColumn(field, v, int8ToColumn)
	},
	serialization.FieldKindArrayOfNullableInt8: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfNullableInt8(field)
		return nullableArrayToColumn(field, TypeInt8Array, vs, int8ToColumn)
	},
	serialization.FieldKindNullableInt16: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadNullableInt16(field)
		return nullableToColumn(field, vs, int16ToColumn)
	},
	serialization.FieldKindArrayOfNullableInt16: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfNullableInt16(field)
		return nullableArrayToColumn(field, TypeInt16Array, vs, int16ToColumn)
	},
	serialization.FieldKindNullableInt32: func(r serialization.CompactReader, field string) Column {
		v := r.ReadNullableInt32(field)
		return nullableToColumn(field, v, int32ToColumn)
	},
	serialization.FieldKindArrayOfNullableInt32: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfNullableInt32(field)
		return nullableArrayToColumn(field, TypeInt32Array, vs, int32ToColumn)
	},
	serialization.FieldKindNullableInt64: func(r serialization.CompactReader, field string) Column {
		v := r.ReadNullableInt64(field)
		return nullableToColumn(field, v, int64ToColumn)
	},
	serialization.FieldKindArrayOfNullableInt64: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfNullableInt64(field)
		return nullableArrayToColumn(field, TypeInt64Array, vs, int64ToColumn)
	},
	serialization.FieldKindNullableFloat32: func(r serialization.CompactReader, field string) Column {
		v := r.ReadNullableFloat32(field)
		return nullableToColumn(field, v, float32ToColumn)
	},
	serialization.FieldKindArrayOfNullableFloat32: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfNullableFloat32(field)
		return nullableArrayToColumn(field, TypeFloat32Array, vs, float32ToColumn)
	},
	serialization.FieldKindNullableFloat64: func(r serialization.CompactReader, field string) Column {
		v := r.ReadNullableFloat64(field)
		return nullableToColumn(field, v, float64ToColumn)
	},
	serialization.FieldKindArrayOfNullableFloat64: func(r serialization.CompactReader, field string) Column {
		vs := r.ReadArrayOfNullableFloat64(field)
		return nullableArrayToColumn(field, TypeFloat64Array, vs, float64ToColumn)
	},
}

func portableToColumn(name string, v serialization.Portable) Column {
	return Column{
		Name:  name,
		Type:  TypePortable,
		Value: v,
	}
}

func byteToColumn(name string, v byte) Column {
	return Column{
		Name:  name,
		Type:  TypeByte,
		Value: v,
	}
}

func uint16ToColumn(name string, v uint16) Column {
	return Column{
		Name:  name,
		Type:  TypeUInt16,
		Value: v,
	}
}

func stringToColumn(name, v string) Column {
	return Column{
		Name:  name,
		Type:  TypeString,
		Value: v,
	}
}

func decimalToColumn(name string, v types.Decimal) Column {
	return Column{
		Name:  name,
		Type:  TypeJavaDecimal,
		Value: v,
	}
}

func localTimeToColumn(name string, v types.LocalTime) Column {
	return Column{
		Name:  name,
		Type:  TypeJavaLocalTime,
		Value: v,
	}
}

func localDateToColumn(name string, v types.LocalDate) Column {
	return Column{
		Name:  name,
		Type:  TypeJavaLocalDate,
		Value: v,
	}
}

func localDateTimeToColumn(name string, v types.LocalDateTime) Column {
	return Column{
		Name:  name,
		Type:  TypeJavaLocalDateTime,
		Value: v,
	}
}

func offsetDateTimeToColumn(name string, v types.OffsetDateTime) Column {
	return Column{
		Name:  name,
		Type:  TypeJavaOffsetDateTime,
		Value: v,
	}
}
