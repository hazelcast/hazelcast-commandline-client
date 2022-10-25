package serialization

import (
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
)

type PortableField struct {
	Name  string
	Type  PortableFieldType
	Value any
}

func (f PortableField) String() string {
	return fmt.Sprintf("%s:%s", f.Name, f.formatValue())
}

func (f PortableField) formatValue() (v string) {
	defer func() {
		if e := recover(); e != nil {
			v = ValueNotDecoded
		}
	}()
	if f.Value == nil {
		return ""
	}
	switch f.Type {
	case PortableTypeNone:
	case PortableTypePortable:
		// nested portable is not supported yet
		return ValueNotDecoded
	case PortableTypeByte, PortableTypeBool, PortableTypeUint16,
		PortableTypeInt16, PortableTypeInt32, PortableTypeInt64,
		PortableTypeFloat32, PortableTypeFloat64, PortableTypeString,
		PortableTypePortableArray, PortableTypeByteArray, PortableTypeBoolArray,
		PortableTypeUInt16Array, PortableTypeInt16Array, PortableTypeInt32Array,
		PortableTypeInt64Array, PortableTypeFloat32Array, PortableTypeFloat64Array,
		PortableTypeStringArray:
		return fmt.Sprintf("%v", f.Value)
	case PortableTypeDecimal:
		return MarshalDecimal(f.Value)
	case PortableTypeDecimalArray:
		return ValueNotDecoded
	case PortableTypeTime:
		sr, err := MarshalLocalTime(f.Value)
		if err != nil {
			return ValueNotDecoded
		} else if sr == nil {
			return ValueNil
		} else {
			return *sr
		}
	case PortableTypeTimeArray:
		return ValueNotDecoded
	case PortableTypeDate:
		sr, err := MarshalLocalDate(f.Value)
		if err != nil {
			return ValueNotDecoded
		} else if sr == nil {
			return ValueNil
		} else {
			return *sr
		}
	case PortableTypeDateArray:
		return ValueNotDecoded
	case PortableTypeTimestamp:
		sr, err := MarshalLocalDateTime(f.Value)
		if err != nil {
			return ValueNotDecoded
		} else if sr == nil {
			return ValueNil
		} else {
			return *sr
		}
	case PortableTypeTimestampArray:
		return ValueNotDecoded
	case PortableTypeTimestampWithTimezone:
		sr, err := MarshalOffsetDateTime(f.Value)
		if err != nil {
			return ValueNotDecoded
		} else if sr == nil {
			return ValueNil
		} else {
			return *sr
		}
	case PortableTypeTimestampWithTimezoneArray:
		return ValueNotDecoded
	default:
		return ValueUnknown
	}
	return v
}

// PortableFieldType corresponds to FieldDefinitionType+1
type PortableFieldType int32

const (
	PortableTypeNone                       PortableFieldType = 0
	PortableTypePortable                   PortableFieldType = 1
	PortableTypeByte                       PortableFieldType = 2
	PortableTypeBool                       PortableFieldType = 3
	PortableTypeUint16                     PortableFieldType = 4
	PortableTypeInt16                      PortableFieldType = 5
	PortableTypeInt32                      PortableFieldType = 6
	PortableTypeInt64                      PortableFieldType = 7
	PortableTypeFloat32                    PortableFieldType = 8
	PortableTypeFloat64                    PortableFieldType = 9
	PortableTypeString                     PortableFieldType = 10
	PortableTypePortableArray              PortableFieldType = 11
	PortableTypeByteArray                  PortableFieldType = 12
	PortableTypeBoolArray                  PortableFieldType = 13
	PortableTypeUInt16Array                PortableFieldType = 14
	PortableTypeInt16Array                 PortableFieldType = 15
	PortableTypeInt32Array                 PortableFieldType = 16
	PortableTypeInt64Array                 PortableFieldType = 17
	PortableTypeFloat32Array               PortableFieldType = 18
	PortableTypeFloat64Array               PortableFieldType = 19
	PortableTypeStringArray                PortableFieldType = 20
	PortableTypeDecimal                    PortableFieldType = 21
	PortableTypeDecimalArray               PortableFieldType = 22
	PortableTypeTime                       PortableFieldType = 23
	PortableTypeTimeArray                  PortableFieldType = 24
	PortableTypeDate                       PortableFieldType = 25
	PortableTypeDateArray                  PortableFieldType = 26
	PortableTypeTimestamp                  PortableFieldType = 27
	PortableTypeTimestampArray             PortableFieldType = 28
	PortableTypeTimestampWithTimezone      PortableFieldType = 29
	PortableTypeTimestampWithTimezoneArray PortableFieldType = 30
)

func (t PortableFieldType) MarshalText() ([]byte, error) {
	var s string
	switch t {
	case PortableTypeNone:
		s = ""
	case PortableTypePortable:
		s = "portable"
	case PortableTypeByte:
		s = "byte"
	case PortableTypeBool:
		s = "bool"
	case PortableTypeUint16:
		s = "uint16"
	case PortableTypeInt16:
		s = "int16"
	case PortableTypeInt32:
		s = "int32"
	case PortableTypeInt64:
		s = "int64"
	case PortableTypeFloat32:
		s = "float32"
	case PortableTypeFloat64:
		s = "float64"
	case PortableTypeString:
		s = "string"
	case PortableTypePortableArray:
		s = "portablearray"
	case PortableTypeByteArray:
		s = "bytearray"
	case PortableTypeBoolArray:
		s = "boolarray"
	case PortableTypeUInt16Array:
		s = "uint16array"
	case PortableTypeInt16Array:
		s = "int16array"
	case PortableTypeInt32Array:
		s = "int32array"
	case PortableTypeInt64Array:
		s = "int64array"
	case PortableTypeFloat32Array:
		s = "float32array"
	case PortableTypeFloat64Array:
		s = "float64array"
	case PortableTypeStringArray:
		s = "stringarray"
	case PortableTypeDecimal:
		s = "decimal"
	case PortableTypeDecimalArray:
		s = "decimalarray"
	case PortableTypeTime:
		s = "time"
	case PortableTypeTimeArray:
		s = "timearray"
	case PortableTypeDate:
		s = "date"
	case PortableTypeDateArray:
		s = "datearray"
	case PortableTypeTimestamp:
		s = "timestamp"
	case PortableTypeTimestampArray:
		s = "timestamparray"
	case PortableTypeTimestampWithTimezone:
		s = "timestampwithtimezone"
	case PortableTypeTimestampWithTimezoneArray:
		s = "timestampwithtimezonearray"
	default:
		return nil, fmt.Errorf("unknown portable type: %d", t)
	}
	return []byte(s), nil
}

func (t *PortableFieldType) UnmarshalText(b []byte) error {
	s := strings.ToLower(string(b))
	switch s {
	case "":
		*t = PortableTypeNone
	case "byte":
		*t = PortableTypePortable
	case "bool":
		*t = PortableTypeBool
	case "uint16":
		*t = PortableTypeUint16
	case "int16":
		*t = PortableTypeInt16
	case "int32":
		*t = PortableTypeInt32
	case "int64":
		*t = PortableTypeInt64
	case "float32":
		*t = PortableTypeFloat32
	case "float64":
		*t = PortableTypeFloat64
	case "string":
		*t = PortableTypeString
	case "portablearray":
		*t = PortableTypePortableArray
	case "bytearray":
		*t = PortableTypeByteArray
	case "boolarray":
		*t = PortableTypeBoolArray
	case "uint16array":
		*t = PortableTypeUInt16Array
	case "int16array":
		*t = PortableTypeInt16Array
	case "int32array":
		*t = PortableTypeInt32Array
	case "int64array":
		*t = PortableTypeInt64Array
	case "float32array":
		*t = PortableTypeFloat32Array
	case "float64array":
		*t = PortableTypeFloat64Array
	case "stringarray":
		*t = PortableTypeStringArray
	case "decimal":
		*t = PortableTypeDecimal
	case "decimalarray":
		*t = PortableTypeDecimalArray
	case "time":
		*t = PortableTypeTime
	case "timearray":
		*t = PortableTypeTimeArray
	case "date":
		*t = PortableTypeDate
	case "datearray":
		*t = PortableTypeDateArray
	case "timestamp":
		*t = PortableTypeTimestamp
	case "timestamparray":
		*t = PortableTypeTimestampArray
	case "timestampwithtimezone":
		*t = PortableTypeTimestampWithTimezone
	case "timestampwithtimezonearray":
		*t = PortableTypeTimestampWithTimezoneArray
	default:
		*t = PortableTypePortable
	}
	return nil
}

func (t *PortableFieldType) ToTypeID() int32 {
	switch *t {
	case PortableTypeNone:
		return TypeNil
	case PortableTypePortable:
		return TypePortable
	case PortableTypeByte:
		return TypeByte
	case PortableTypeBool:
		return TypeBool
	case PortableTypeUint16:
		return TypeUInt16
	case PortableTypeInt16:
		return TypeInt16
	case PortableTypeInt32:
		return TypeInt32
	case PortableTypeInt64:
		return TypeInt64
	case PortableTypeFloat32:
		return TypeFloat32
	case PortableTypeFloat64:
		return TypeFloat64
	case PortableTypeString:
		return TypeString
	case PortableTypePortableArray:
		return TypeJavaArray
	case PortableTypeByteArray:
		return TypeByteArray
	case PortableTypeBoolArray:
		return TypeBoolArray
	case PortableTypeUInt16Array:
		return TypeUInt16Array
	case PortableTypeInt16Array:
		return TypeInt16Array
	case PortableTypeInt32Array:
		return TypeInt32Array
	case PortableTypeInt64Array:
		return TypeInt64Array
	case PortableTypeFloat32Array:
		return TypeFloat32Array
	case PortableTypeFloat64Array:
		return TypeFloat64Array
	case PortableTypeStringArray:
		return TypeStringArray
	case PortableTypeDecimal:
		return TypeJavaDecimal
	case PortableTypeDecimalArray:
		return TypeJavaArray
	case PortableTypeTime:
		return TypeJavaLocalTime
	case PortableTypeTimeArray:
		return TypeJavaArray
	case PortableTypeDate:
		return TypeJavaLocalDate
	case PortableTypeDateArray:
		return TypeJavaArray
	case PortableTypeTimestamp:
		return TypeJavaLocalDateTime
	case PortableTypeTimestampArray:
		return TypeJavaArray
	case PortableTypeTimestampWithTimezone:
		return TypeJavaOffsetDateTime
	case PortableTypeTimestampWithTimezoneArray:
		return TypeJavaArray
	}
	return TypeUnknown
}

var portableReaders = map[serialization.FieldDefinitionType]portableFieldReader{
	serialization.TypeBool: func(r serialization.PortableReader, field string) any {
		return r.ReadBool(field)
	},
	serialization.TypeString: func(r serialization.PortableReader, field string) any {
		return r.ReadString(field)
	},
	serialization.TypeDate: func(r serialization.PortableReader, field string) any {
		return r.ReadDate(field)
	},
	serialization.TypeTime: func(r serialization.PortableReader, field string) any {
		return r.ReadTime(field)
	},
	serialization.TypeTimestamp: func(r serialization.PortableReader, field string) any {
		return r.ReadTimestamp(field)
	},
	serialization.TypeTimestampWithTimezone: func(r serialization.PortableReader, field string) any {
		return r.ReadTimestampWithTimezone(field)
	},
	serialization.TypeInt32: func(r serialization.PortableReader, field string) any {
		return r.ReadInt32(field)
	},
}

// the list of writers is not complete, but that's OK since we don't yet support writing portables --YT
var portableWriters = map[serialization.FieldDefinitionType]portableFieldWriter{
	serialization.TypeBool: func(w serialization.PortableWriter, field string, value any) {
		w.WriteBool(field, value.(bool))
	},
	serialization.TypeString: func(w serialization.PortableWriter, field string, value any) {
		w.WriteString(field, value.(string))
	},
	serialization.TypeDate: func(w serialization.PortableWriter, field string, value any) {
		w.WriteDate(field, value.(*types.LocalDate))
	},
	serialization.TypeInt32: func(w serialization.PortableWriter, field string, value any) {
		w.WriteInt32(field, value.(int32))
	},
}
