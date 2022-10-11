package serialization

import (
	"fmt"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
)

type FieldType int

type Field struct {
	Name  string
	Type  PortableType
	Value any
}

func (f Field) String() string {
	return fmt.Sprintf("%s:%s", f.Name, f.formatValue())
}

func (f Field) formatValue() (v string) {
	defer func() {
		if e := recover(); e != nil {
			v = "[ERROR]"
		}
	}()
	if f.Value == nil {
		return ""
	}
	switch f.Type {
	case PortableTypeNone:
	case PortableTypePortable:
		return "[PORTABLE]"
	case PortableTypeByte:
		fallthrough
	case PortableTypeBool:
		fallthrough
	case PortableTypeUint16:
		fallthrough
	case PortableTypeInt16:
		fallthrough
	case PortableTypeInt32:
		fallthrough
	case PortableTypeInt64:
		fallthrough
	case PortableTypeFloat32:
		fallthrough
	case PortableTypeFloat64:
		fallthrough
	case PortableTypeString:
		fallthrough
	case PortableTypePortableArray:
		fallthrough
	case PortableTypeByteArray:
		fallthrough
	case PortableTypeBoolArray:
		fallthrough
	case PortableTypeUInt16Array:
		fallthrough
	case PortableTypeInt16Array:
		fallthrough
	case PortableTypeInt32Array:
		fallthrough
	case PortableTypeInt64Array:
		fallthrough
	case PortableTypeFloat32Array:
		fallthrough
	case PortableTypeFloat64Array:
		fallthrough
	case PortableTypeStringArray:
		fallthrough
	case PortableTypeDecimal:
		fallthrough
	case PortableTypeDecimalArray:
		v = fmt.Sprintf("%v", f.Value)
	case PortableTypeTime:
		v = (*time.Time)(f.Value.(*types.LocalTime)).Format("15:04:05")
	case PortableTypeTimeArray:
		v = "timearray"
	case PortableTypeDate:
		v = (*time.Time)(f.Value.(*types.LocalDate)).Format("2006-01-02")
	case PortableTypeDateArray:
		v = "datearray"
	case PortableTypeTimestamp:
		v = (*time.Time)(f.Value.(*types.LocalDateTime)).Format("2006-01-02 15:04:05")
	case PortableTypeTimestampArray:
		v = "timestamparray"
	case PortableTypeTimestampWithTimezone:
		v = (*time.Time)(f.Value.(*types.OffsetDateTime)).Format(time.RFC3339)
	case PortableTypeTimestampWithTimezoneArray:
		v = "timestampwithtimezonearray"
	default:
		v = "[UNKNOWN]"
	}
	return v
}

// PortableType corresponds to FieldDefinitionType+1
type PortableType int32

const (
	PortableTypeNone                       PortableType = 0
	PortableTypePortable                   PortableType = 1
	PortableTypeByte                       PortableType = 2
	PortableTypeBool                       PortableType = 3
	PortableTypeUint16                     PortableType = 4
	PortableTypeInt16                      PortableType = 5
	PortableTypeInt32                      PortableType = 6
	PortableTypeInt64                      PortableType = 7
	PortableTypeFloat32                    PortableType = 8
	PortableTypeFloat64                    PortableType = 9
	PortableTypeString                     PortableType = 10
	PortableTypePortableArray              PortableType = 11
	PortableTypeByteArray                  PortableType = 12
	PortableTypeBoolArray                  PortableType = 13
	PortableTypeUInt16Array                PortableType = 14
	PortableTypeInt16Array                 PortableType = 15
	PortableTypeInt32Array                 PortableType = 16
	PortableTypeInt64Array                 PortableType = 17
	PortableTypeFloat32Array               PortableType = 18
	PortableTypeFloat64Array               PortableType = 19
	PortableTypeStringArray                PortableType = 20
	PortableTypeDecimal                    PortableType = 21
	PortableTypeDecimalArray               PortableType = 22
	PortableTypeTime                       PortableType = 23
	PortableTypeTimeArray                  PortableType = 24
	PortableTypeDate                       PortableType = 25
	PortableTypeDateArray                  PortableType = 26
	PortableTypeTimestamp                  PortableType = 27
	PortableTypeTimestampArray             PortableType = 28
	PortableTypeTimestampWithTimezone      PortableType = 29
	PortableTypeTimestampWithTimezoneArray PortableType = 30
)

func (t PortableType) MarshalText() ([]byte, error) {
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

func (t *PortableType) UnmarshalText(b []byte) error {
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
	serialization.TypeInt32: func(r serialization.PortableReader, field string) any {
		return r.ReadInt32(field)
	},
}

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
