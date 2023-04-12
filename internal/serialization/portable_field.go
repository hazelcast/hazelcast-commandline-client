package serialization

import (
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

var FieldDefinitionIDToType = map[serialization.FieldDefinitionType]int32{
	serialization.TypePortable:                   TypePortable,
	serialization.TypeByte:                       TypeByte,
	serialization.TypeBool:                       TypeBool,
	serialization.TypeUint16:                     TypeUInt16,
	serialization.TypeInt16:                      TypeInt16,
	serialization.TypeInt32:                      TypeInt32,
	serialization.TypeInt64:                      TypeInt64,
	serialization.TypeFloat32:                    TypeFloat32,
	serialization.TypeFloat64:                    TypeFloat64,
	serialization.TypeString:                     TypeString,
	serialization.TypePortableArray:              TypeJavaArray,
	serialization.TypeByteArray:                  TypeByteArray,
	serialization.TypeBoolArray:                  TypeBoolArray,
	serialization.TypeUInt16Array:                TypeUInt16Array,
	serialization.TypeInt16Array:                 TypeInt16Array,
	serialization.TypeInt32Array:                 TypeInt32Array,
	serialization.TypeInt64Array:                 TypeInt64Array,
	serialization.TypeFloat32Array:               TypeFloat32Array,
	serialization.TypeFloat64Array:               TypeFloat64Array,
	serialization.TypeStringArray:                TypeStringArray,
	serialization.TypeDecimal:                    TypeJavaDecimal,
	serialization.TypeDecimalArray:               TypeDecimalArray,
	serialization.TypeTime:                       TypeJavaLocalTime,
	serialization.TypeTimeArray:                  TypeJavaLocalTimeArray,
	serialization.TypeDate:                       TypeJavaLocalDate,
	serialization.TypeDateArray:                  TypeJavaLocalDateArray,
	serialization.TypeTimestamp:                  TypeJavaLocalDateTime,
	serialization.TypeTimestampArray:             TypeJavaLocalDateTimeArray,
	serialization.TypeTimestampWithTimezone:      TypeJavaOffsetDateTime,
	serialization.TypeTimestampWithTimezoneArray: TypeJavaOffsetDateTimeArray,
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
