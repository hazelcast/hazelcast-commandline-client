package serialization

import (
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/serialization"
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

var portableFieldTypeToString = map[PortableFieldType]string{
	PortableTypeNone:                       "",
	PortableTypePortable:                   "portable",
	PortableTypeByte:                       "byte",
	PortableTypeBool:                       "bool",
	PortableTypeUint16:                     "uint16",
	PortableTypeInt16:                      "int16",
	PortableTypeInt32:                      "int32",
	PortableTypeInt64:                      "int64",
	PortableTypeFloat32:                    "float32",
	PortableTypeFloat64:                    "float64",
	PortableTypeString:                     "string",
	PortableTypePortableArray:              "portablearray",
	PortableTypeByteArray:                  "bytearray",
	PortableTypeBoolArray:                  "boolarray",
	PortableTypeUInt16Array:                "uint16array",
	PortableTypeInt16Array:                 "int16array",
	PortableTypeInt32Array:                 "int32array",
	PortableTypeInt64Array:                 "int64array",
	PortableTypeFloat32Array:               "float32array",
	PortableTypeFloat64Array:               "float64array",
	PortableTypeStringArray:                "stringarray",
	PortableTypeDecimal:                    "decimal",
	PortableTypeDecimalArray:               "decimalarray",
	PortableTypeTime:                       "time",
	PortableTypeTimeArray:                  "timearray",
	PortableTypeDate:                       "date",
	PortableTypeDateArray:                  "datearray",
	PortableTypeTimestamp:                  "timestamp",
	PortableTypeTimestampArray:             "timestamparray",
	PortableTypeTimestampWithTimezone:      "timestampwithtimezone",
	PortableTypeTimestampWithTimezoneArray: "timestampwithtimezonearray",
}

func (t PortableFieldType) MarshalText() ([]byte, error) {
	s, ok := portableFieldTypeToString[t]
	if !ok {
		return nil, fmt.Errorf("unknown portable type: %d", t)
	}
	return []byte(s), nil
}

var stringToPortableFieldType = map[string]PortableFieldType{
	"":                           PortableTypeNone,
	"byte":                       PortableTypePortable,
	"bool":                       PortableTypeBool,
	"uint16":                     PortableTypeUint16,
	"int16":                      PortableTypeInt16,
	"int32":                      PortableTypeInt32,
	"int64":                      PortableTypeInt64,
	"float32":                    PortableTypeFloat32,
	"float64":                    PortableTypeFloat64,
	"string":                     PortableTypeString,
	"portablearray":              PortableTypePortableArray,
	"bytearray":                  PortableTypeByteArray,
	"boolarray":                  PortableTypeBoolArray,
	"uint16array":                PortableTypeUInt16Array,
	"int16array":                 PortableTypeInt16Array,
	"int32array":                 PortableTypeInt32Array,
	"int64array":                 PortableTypeInt64Array,
	"float32array":               PortableTypeFloat32Array,
	"float64array":               PortableTypeFloat64Array,
	"stringarray":                PortableTypeStringArray,
	"decimal":                    PortableTypeDecimal,
	"decimalarray":               PortableTypeDecimalArray,
	"time":                       PortableTypeTime,
	"timearray":                  PortableTypeTimeArray,
	"date":                       PortableTypeDate,
	"datearray":                  PortableTypeDateArray,
	"timestamp":                  PortableTypeTimestamp,
	"timestamparray":             PortableTypeTimestampArray,
	"timestampwithtimezone":      PortableTypeTimestampWithTimezone,
	"timestampwithtimezonearray": PortableTypeTimestampWithTimezoneArray,
}

func (t *PortableFieldType) UnmarshalText(b []byte) error {
	ft, ok := stringToPortableFieldType[strings.ToLower(string(b))]
	if !ok {
		ft = PortableTypePortable
	}
	*t = ft
	return nil
}

var portableFieldTypeToTypeID = map[PortableFieldType]int32{
	PortableTypeNone:                       TypeNil,
	PortableTypePortable:                   TypePortable,
	PortableTypeByte:                       TypeByte,
	PortableTypeBool:                       TypeBool,
	PortableTypeUint16:                     TypeUInt16,
	PortableTypeInt16:                      TypeInt16,
	PortableTypeInt32:                      TypeInt32,
	PortableTypeInt64:                      TypeInt64,
	PortableTypeFloat32:                    TypeFloat32,
	PortableTypeFloat64:                    TypeFloat64,
	PortableTypeString:                     TypeString,
	PortableTypePortableArray:              TypeJavaArray,
	PortableTypeByteArray:                  TypeByteArray,
	PortableTypeBoolArray:                  TypeBoolArray,
	PortableTypeUInt16Array:                TypeUInt16Array,
	PortableTypeInt16Array:                 TypeInt16Array,
	PortableTypeInt32Array:                 TypeInt32Array,
	PortableTypeInt64Array:                 TypeInt64Array,
	PortableTypeFloat32Array:               TypeFloat32Array,
	PortableTypeFloat64Array:               TypeFloat64Array,
	PortableTypeStringArray:                TypeStringArray,
	PortableTypeDecimal:                    TypeJavaDecimal,
	PortableTypeDecimalArray:               TypeJavaArray,
	PortableTypeTime:                       TypeJavaLocalTime,
	PortableTypeTimeArray:                  TypeJavaArray,
	PortableTypeDate:                       TypeJavaLocalDate,
	PortableTypeDateArray:                  TypeJavaArray,
	PortableTypeTimestamp:                  TypeJavaLocalDateTime,
	PortableTypeTimestampArray:             TypeJavaArray,
	PortableTypeTimestampWithTimezone:      TypeJavaOffsetDateTime,
	PortableTypeTimestampWithTimezoneArray: TypeJavaArray,
}

func (t *PortableFieldType) ToTypeID() int32 {
	id, ok := portableFieldTypeToTypeID[*t]
	if !ok {
		return TypeUnknown
	}
	return id

}

var portableReaders = map[serialization.FieldDefinitionType]portableFieldReader{
	serialization.TypePortable: func(r serialization.PortableReader, field string) any {
		return r.ReadPortable(field)
	},
	serialization.TypeByte: func(r serialization.PortableReader, field string) any {
		return r.ReadByte(field)
	},
	serialization.TypeBool: func(r serialization.PortableReader, field string) any {
		return r.ReadBool(field)
	},
	serialization.TypeUint16: func(r serialization.PortableReader, field string) any {
		return r.ReadUInt16(field)
	},
	serialization.TypeInt16: func(r serialization.PortableReader, field string) any {
		return r.ReadInt16(field)
	},
	serialization.TypeInt32: func(r serialization.PortableReader, field string) any {
		return r.ReadInt32(field)
	},
	serialization.TypeInt64: func(r serialization.PortableReader, field string) any {
		return r.ReadInt64(field)
	},
	serialization.TypeFloat32: func(r serialization.PortableReader, field string) any {
		return r.ReadFloat32(field)
	},
	serialization.TypeFloat64: func(r serialization.PortableReader, field string) any {
		return r.ReadFloat64(field)
	},
	serialization.TypeString: func(r serialization.PortableReader, field string) any {
		return r.ReadString(field)
	},
	serialization.TypePortableArray: func(r serialization.PortableReader, field string) any {
		return r.ReadPortableArray(field)
	},
	serialization.TypeByteArray: func(r serialization.PortableReader, field string) any {
		return r.ReadByteArray(field)
	},
	serialization.TypeBoolArray: func(r serialization.PortableReader, field string) any {
		return r.ReadBoolArray(field)
	},
	serialization.TypeUInt16Array: func(r serialization.PortableReader, field string) any {
		return r.ReadUInt16Array(field)
	},
	serialization.TypeInt16Array: func(r serialization.PortableReader, field string) any {
		return r.ReadInt16Array(field)
	},
	serialization.TypeInt32Array: func(r serialization.PortableReader, field string) any {
		return r.ReadInt32Array(field)
	},
	serialization.TypeInt64Array: func(r serialization.PortableReader, field string) any {
		return r.ReadInt64Array(field)
	},
	serialization.TypeFloat32Array: func(r serialization.PortableReader, field string) any {
		return r.ReadFloat32Array(field)
	},
	serialization.TypeFloat64Array: func(r serialization.PortableReader, field string) any {
		return r.ReadFloat64Array(field)
	},
	serialization.TypeStringArray: func(r serialization.PortableReader, field string) any {
		return r.ReadStringArray(field)
	},
	serialization.TypeDecimal: func(r serialization.PortableReader, field string) any {
		return r.ReadDecimal(field)
	},
	serialization.TypeDecimalArray: func(r serialization.PortableReader, field string) any {
		return r.ReadDecimalArray(field)
	},
	serialization.TypeTime: func(r serialization.PortableReader, field string) any {
		return r.ReadTime(field)
	},
	serialization.TypeTimeArray: func(r serialization.PortableReader, field string) any {
		return r.ReadTimeArray(field)
	},
	serialization.TypeDate: func(r serialization.PortableReader, field string) any {
		return r.ReadDate(field)
	},
	serialization.TypeDateArray: func(r serialization.PortableReader, field string) any {
		return r.ReadDateArray(field)
	},
	serialization.TypeTimestamp: func(r serialization.PortableReader, field string) any {
		return r.ReadTimestamp(field)
	},
	serialization.TypeTimestampArray: func(r serialization.PortableReader, field string) any {
		return r.ReadTimestampArray(field)
	},
	serialization.TypeTimestampWithTimezone: func(r serialization.PortableReader, field string) any {
		return r.ReadTimestampWithTimezone(field)
	},
	serialization.TypeTimestampWithTimezoneArray: func(r serialization.PortableReader, field string) any {
		return r.ReadTimestampWithTimezoneArray(field)
	},
}
