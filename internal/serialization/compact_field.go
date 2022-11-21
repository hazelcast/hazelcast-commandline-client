package serialization

import (
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type CompactField struct {
	Name  string
	Type  CompactFieldType
	Value any
}

func readNullableCompactField[T any](field string, f func(string) *T) any {
	v := f(field)
	if v == nil {
		return nil
	}
	return *v
}

var compactReaders = map[serialization.FieldKind]compactFieldReader{
	serialization.FieldKindBoolean: func(r serialization.CompactReader, field string) any {
		return r.ReadBoolean(field)
	},
	serialization.FieldKindArrayOfBoolean: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfBoolean(field)
	},
	serialization.FieldKindInt8: func(r serialization.CompactReader, field string) any {
		return r.ReadInt8(field)
	},
	serialization.FieldKindArrayOfInt8: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfInt8(field)
	},
	// FieldKindChar        : Not decoded due to spec.
	// FieldKindArrayOfChar : Not decoded due to spec.
	serialization.FieldKindInt16: func(r serialization.CompactReader, field string) any {
		return r.ReadInt16(field)
	},
	serialization.FieldKindArrayOfInt16: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfInt16(field)
	},
	serialization.FieldKindInt32: func(r serialization.CompactReader, field string) any {
		return r.ReadInt32(field)
	},
	serialization.FieldKindArrayOfInt32: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfInt32(field)
	},
	serialization.FieldKindInt64: func(r serialization.CompactReader, field string) any {
		return r.ReadInt64(field)
	},
	serialization.FieldKindArrayOfInt64: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfInt64(field)
	},
	serialization.FieldKindFloat32: func(r serialization.CompactReader, field string) any {
		return r.ReadFloat32(field)
	},
	serialization.FieldKindArrayOfFloat32: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfFloat32(field)
	},
	serialization.FieldKindFloat64: func(r serialization.CompactReader, field string) any {
		return r.ReadFloat64(field)
	},
	serialization.FieldKindArrayOfFloat64: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfFloat64(field)
	},
	serialization.FieldKindString: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadString)
	},
	serialization.FieldKindArrayOfString: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfString(field)
	},
	serialization.FieldKindDecimal: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadDecimal)
	},
	serialization.FieldKindArrayOfDecimal: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfDecimal(field)
	},
	serialization.FieldKindTime: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadTime)
	},
	serialization.FieldKindArrayOfTime: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfTime(field)
	},
	serialization.FieldKindDate: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadDate)
	},
	serialization.FieldKindArrayOfDate: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfDate(field)
	},
	serialization.FieldKindTimestamp: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadTimestamp)
	},
	serialization.FieldKindArrayOfTimestamp: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfTimestamp(field)
	},
	serialization.FieldKindTimestampWithTimezone: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadTimestampWithTimezone)
	},
	serialization.FieldKindArrayOfTimestampWithTimezone: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfTimestampWithTimezone(field)
	},
	serialization.FieldKindCompact: func(r serialization.CompactReader, field string) any {
		return r.ReadCompact(field)
	},
	serialization.FieldKindArrayOfCompact: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfCompact(field)
	},
	// FieldKindPortable        : Not decoded due to spec.
	// FieldKindArrayOfPortable : Not decoded due to spec.
	serialization.FieldKindNullableBoolean: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadNullableBoolean)
	},
	serialization.FieldKindArrayOfNullableBoolean: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfNullableBoolean(field)
	},
	serialization.FieldKindNullableInt8: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadNullableInt8)
	},
	serialization.FieldKindArrayOfNullableInt8: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfNullableInt8(field)
	},
	serialization.FieldKindNullableInt16: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadNullableInt16)
	},
	serialization.FieldKindArrayOfNullableInt16: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfNullableInt16(field)
	},
	serialization.FieldKindNullableInt32: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadNullableInt32)
	},
	serialization.FieldKindArrayOfNullableInt32: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfNullableInt32(field)
	},
	serialization.FieldKindNullableInt64: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadNullableInt64)
	},
	serialization.FieldKindArrayOfNullableInt64: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfNullableInt64(field)
	},
	serialization.FieldKindNullableFloat32: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadNullableFloat32)
	},
	serialization.FieldKindArrayOfNullableFloat32: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfNullableFloat32(field)
	},
	serialization.FieldKindNullableFloat64: func(r serialization.CompactReader, field string) any {
		return readNullableCompactField(field, r.ReadNullableFloat64)
	},
	serialization.FieldKindArrayOfNullableFloat64: func(r serialization.CompactReader, field string) any {
		return r.ReadArrayOfNullableFloat64(field)
	},
}

type CompactFieldType serialization.FieldKind

var stringToCompactFieldType = map[string]serialization.FieldKind{
	"":           serialization.FieldKindNotAvailable,
	"bool":       serialization.FieldKindBoolean,
	"bool_array": serialization.FieldKindArrayOfBoolean,
	"int8":       serialization.FieldKindInt8,
	"int8_array": serialization.FieldKindArrayOfInt8,
	// FieldKindChar        : Not available due to spec.
	// FieldKindArrayOfChar : Not available due to spec.
	"int16":                         serialization.FieldKindInt16,
	"int16_array":                   serialization.FieldKindArrayOfInt16,
	"int32":                         serialization.FieldKindInt32,
	"int32_array":                   serialization.FieldKindArrayOfInt32,
	"int64":                         serialization.FieldKindInt64,
	"int64_array":                   serialization.FieldKindArrayOfInt64,
	"float32":                       serialization.FieldKindFloat32,
	"float32_array":                 serialization.FieldKindArrayOfFloat32,
	"float64":                       serialization.FieldKindFloat64,
	"float64_array":                 serialization.FieldKindArrayOfFloat64,
	"string":                        serialization.FieldKindString,
	"string_array":                  serialization.FieldKindArrayOfString,
	"decimal":                       serialization.FieldKindDecimal,
	"decimal_array":                 serialization.FieldKindArrayOfDecimal,
	"time":                          serialization.FieldKindTime,
	"time_array":                    serialization.FieldKindArrayOfTime,
	"date":                          serialization.FieldKindDate,
	"date_array":                    serialization.FieldKindArrayOfDate,
	"timestamp":                     serialization.FieldKindTimestamp,
	"timestamp_array":               serialization.FieldKindArrayOfTimestamp,
	"timestamp_with_timezone":       serialization.FieldKindTimestampWithTimezone,
	"timestamp_with_timezone_array": serialization.FieldKindArrayOfTimestampWithTimezone,
	"compact":                       serialization.FieldKindCompact,
	"compact_array":                 serialization.FieldKindArrayOfCompact,
	// FieldKindPortable        : Not available due to spec.
	// FieldKindArrayOfPortable : Not available due to spec.
	"nullable_bool":          serialization.FieldKindNullableBoolean,
	"nullable_bool_array":    serialization.FieldKindArrayOfNullableBoolean,
	"nullable_int8":          serialization.FieldKindNullableInt8,
	"nullable_int8_array":    serialization.FieldKindArrayOfNullableInt8,
	"nullable_int16":         serialization.FieldKindNullableInt16,
	"nullable_int16_array":   serialization.FieldKindArrayOfNullableInt16,
	"nullable_int32":         serialization.FieldKindNullableInt32,
	"nullable_int32_array":   serialization.FieldKindArrayOfNullableInt32,
	"nullable_int64":         serialization.FieldKindNullableInt64,
	"nullable_int64_array":   serialization.FieldKindArrayOfNullableInt64,
	"nullable_float32":       serialization.FieldKindNullableFloat32,
	"nullable_float32_array": serialization.FieldKindArrayOfNullableFloat32,
	"nullable_float64":       serialization.FieldKindNullableFloat64,
	"nullable_float64_array": serialization.FieldKindArrayOfNullableFloat64,
}

func (t *CompactFieldType) UnmarshalText(b []byte) error {
	s := strings.ToLower(string(b))
	ft, ok := stringToCompactFieldType[s]
	if !ok {
		return fmt.Errorf("unknown type: %s", s)
	}
	*t = CompactFieldType(ft)
	return nil
}
