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
	serialization.TypePortable: func(r serialization.PortableReader, field string) Column {
		return portableToColumn(field, r.ReadPortable(field))
	},
	serialization.TypeByte: func(r serialization.PortableReader, field string) Column {
		return byteToColumn(field, r.ReadByte(field))
	},
	serialization.TypeBool: func(r serialization.PortableReader, field string) Column {
		return boolToColumn(field, r.ReadBool(field))
	},
	serialization.TypeUint16: func(r serialization.PortableReader, field string) Column {
		return uint16ToColumn(field, r.ReadUInt16(field))
	},
	serialization.TypeInt16: func(r serialization.PortableReader, field string) Column {
		return int16ToColumn(field, r.ReadInt16(field))
	},
	serialization.TypeInt32: func(r serialization.PortableReader, field string) Column {
		return int32ToColumn(field, r.ReadInt32(field))
	},
	serialization.TypeInt64: func(r serialization.PortableReader, field string) Column {
		return int64ToColumn(field, r.ReadInt64(field))
	},
	serialization.TypeFloat32: func(r serialization.PortableReader, field string) Column {
		return float32ToColumn(field, r.ReadFloat32(field))
	},
	serialization.TypeFloat64: func(r serialization.PortableReader, field string) Column {
		return float64ToColumn(field, r.ReadFloat64(field))
	},
	serialization.TypeString: func(r serialization.PortableReader, field string) Column {
		return stringToColumn(field, r.ReadString(field))
	},
	serialization.TypePortableArray: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadPortableArray(field)
		return arrayToColumn(field, TypePortableArray, vs, portableToColumn)
	},
	serialization.TypeByteArray: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadByteArray(field)
		return arrayToColumn(field, TypeByteArray, vs, byteToColumn)
	},
	serialization.TypeBoolArray: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadBoolArray(field)
		return arrayToColumn(field, TypeBoolArray, vs, boolToColumn)
	},
	serialization.TypeUInt16Array: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadUInt16Array(field)
		return arrayToColumn(field, TypeUInt16Array, vs, uint16ToColumn)
	},
	serialization.TypeInt16Array: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadInt16Array(field)
		return arrayToColumn(field, TypeInt16Array, vs, int16ToColumn)
	},
	serialization.TypeInt32Array: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadInt32Array(field)
		return arrayToColumn(field, TypeInt32Array, vs, int32ToColumn)
	},
	serialization.TypeInt64Array: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadInt64Array(field)
		return arrayToColumn(field, TypeInt64Array, vs, int64ToColumn)
	},
	serialization.TypeFloat32Array: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadFloat32Array(field)
		return arrayToColumn(field, TypeFloat32Array, vs, float32ToColumn)
	},
	serialization.TypeFloat64Array: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadFloat64Array(field)
		return arrayToColumn(field, TypeFloat64Array, vs, float64ToColumn)
	},
	serialization.TypeStringArray: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadStringArray(field)
		return arrayToColumn(field, TypeStringArray, vs, stringToColumn)
	},
	serialization.TypeDecimal: func(r serialization.PortableReader, field string) Column {
		return ptrDecimalToColumn(field, r.ReadDecimal(field))
	},
	serialization.TypeDecimalArray: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadDecimalArray(field)
		return arrayToColumn(field, TypeDecimalArray, vs, decimalToColumn)
	},
	serialization.TypeTime: func(r serialization.PortableReader, field string) Column {
		return ptrLocalTimeToColumn(field, r.ReadTime(field))
	},
	serialization.TypeTimeArray: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadTimeArray(field)
		return arrayToColumn(field, TypeJavaLocalTimeArray, vs, localTimeToColumn)
	},
	serialization.TypeDate: func(r serialization.PortableReader, field string) Column {
		return ptrLocalDateToColumn(field, r.ReadDate(field))
	},
	serialization.TypeDateArray: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadDateArray(field)
		return arrayToColumn(field, TypeJavaLocalDateArray, vs, localDateToColumn)
	},
	serialization.TypeTimestamp: func(r serialization.PortableReader, field string) Column {
		return ptrLocalDateTimeToColumn(field, r.ReadTimestamp(field))
	},
	serialization.TypeTimestampArray: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadTimestampArray(field)
		return arrayToColumn(field, TypeJavaLocalDateTimeArray, vs, localDateTimeToColumn)
	},
	serialization.TypeTimestampWithTimezone: func(r serialization.PortableReader, field string) Column {
		return ptrOffsetDateTimeToColumn(field, r.ReadTimestampWithTimezone(field))
	},
	serialization.TypeTimestampWithTimezoneArray: func(r serialization.PortableReader, field string) Column {
		vs := r.ReadTimestampWithTimezoneArray(field)
		return arrayToColumn(field, TypeJavaOffsetDateTime, vs, offsetDateTimeToColumn)
	},
}
