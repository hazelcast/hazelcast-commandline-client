package serialization

import (
	"reflect"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type compactFieldReader func(r serialization.CompactReader, field string) any

type SchemaInfo struct {
	Type     reflect.Type
	TypeName string
}

type GenericCompactDeserializer struct{}

func (cm GenericCompactDeserializer) Read(schema *hazelcast.Schema, reader serialization.CompactReader) interface{} {
	fds := schema.FieldDefinitions()
	cs := make(ColumnList, len(fds))
	for i, fd := range fds {
		r := compactReaders[fd.Kind]
		v := r(reader, fd.Name)
		c := Column{
			Name:  fd.Name,
			Value: v,
		}
		if v == nil {
			c.Type = TypeNil
		} else {
			c.Type = fieldKindToType[fd.Kind]
		}
		cs[i] = c
	}
	return cs
}

var fieldKindToType map[serialization.FieldKind]int32

func init() {
	fieldKindToType = map[serialization.FieldKind]int32{
		serialization.FieldKindNotAvailable:                 TypeNil,
		serialization.FieldKindBoolean:                      TypeBool,
		serialization.FieldKindArrayOfBoolean:               TypeBoolArray,
		serialization.FieldKindInt8:                         TypeInt8,
		serialization.FieldKindArrayOfInt8:                  TypeInt8Array,
		serialization.FieldKindInt16:                        TypeInt16,
		serialization.FieldKindArrayOfInt16:                 TypeInt16Array,
		serialization.FieldKindInt32:                        TypeInt32,
		serialization.FieldKindArrayOfInt32:                 TypeInt32Array,
		serialization.FieldKindInt64:                        TypeInt64,
		serialization.FieldKindArrayOfInt64:                 TypeInt64Array,
		serialization.FieldKindFloat32:                      TypeFloat32,
		serialization.FieldKindArrayOfFloat32:               TypeFloat32Array,
		serialization.FieldKindFloat64:                      TypeFloat64,
		serialization.FieldKindArrayOfFloat64:               TypeFloat64Array,
		serialization.FieldKindString:                       TypeString,
		serialization.FieldKindArrayOfString:                TypeStringArray,
		serialization.FieldKindDecimal:                      TypeJavaDecimal,
		serialization.FieldKindArrayOfDecimal:               TypeDecimalArray,
		serialization.FieldKindTime:                         TypeJavaLocalTime,
		serialization.FieldKindArrayOfTime:                  TypeJavaLocalTimeArray,
		serialization.FieldKindDate:                         TypeJavaLocalDate,
		serialization.FieldKindArrayOfDate:                  TypeJavaLocalDateArray,
		serialization.FieldKindTimestamp:                    TypeJavaLocalDateTime,
		serialization.FieldKindArrayOfTimestamp:             TypeJavaLocalDateTimeArray,
		serialization.FieldKindTimestampWithTimezone:        TypeJavaOffsetDateTime,
		serialization.FieldKindArrayOfTimestampWithTimezone: TypeJavaOffsetDateTimeArray,
		serialization.FieldKindCompact:                      TypeCompact,
		serialization.FieldKindArrayOfCompact:               TypeCompactArray,
		serialization.FieldKindNullableBoolean:              TypeBool,
		serialization.FieldKindArrayOfNullableBoolean:       TypeBoolArray,
		serialization.FieldKindNullableInt8:                 TypeInt8,
		serialization.FieldKindArrayOfNullableInt8:          TypeInt8Array,
		serialization.FieldKindNullableInt16:                TypeInt16,
		serialization.FieldKindArrayOfNullableInt16:         TypeInt16Array,
		serialization.FieldKindNullableInt32:                TypeInt32,
		serialization.FieldKindArrayOfNullableInt32:         TypeInt32Array,
		serialization.FieldKindNullableInt64:                TypeInt64,
		serialization.FieldKindArrayOfNullableInt64:         TypeInt64Array,
		serialization.FieldKindNullableFloat32:              TypeFloat32,
		serialization.FieldKindArrayOfNullableFloat32:       TypeFloat32Array,
		serialization.FieldKindNullableFloat64:              TypeFloat64,
		serialization.FieldKindArrayOfNullableFloat64:       TypeFloat64Array,
	}
}
