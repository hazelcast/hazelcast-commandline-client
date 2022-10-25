package serialization

import (
	"fmt"
	"reflect"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type compactFieldReader func(r serialization.CompactReader, field string) any

type compactFieldWriter func(w serialization.CompactWriter, field string, value any)

type GenericCompact struct {
	ValueType     reflect.Type
	ValueTypeName string
	Fields        []CompactField
	readers       []compactFieldReader
	writers       []compactFieldWriter
}

func NewGenericCompact(value GenericCompact) (*GenericCompact, error) {
	rs := make([]compactFieldReader, len(value.Fields))
	ws := make([]compactFieldWriter, len(value.Fields))
	for i, f := range value.Fields {
		if f.Type < CompactFieldType(serialization.FieldKindNotAvailable) || f.Type > CompactFieldType(serialization.FieldKindArrayOfNullableFloat64) {
			return nil, fmt.Errorf("invalid portable type: %d", f.Type)
		}
		r, ok := compactReaders[serialization.FieldKind(f.Type)]
		if !ok {
			return nil, fmt.Errorf("reader not found for compact type: %d", f.Type)
		}
		rs[i] = r
		// writing is disabled for now --YT
	}
	return &GenericCompact{
		Fields:        value.Fields,
		ValueType:     value.makeType(),
		ValueTypeName: value.ValueTypeName,
		readers:       rs,
		writers:       ws,
	}, nil

}

func (cm GenericCompact) makeType() reflect.Type {
	fs := make([]reflect.StructField, len(cm.Fields))
	for i, f := range cm.Fields {
		fs[i] = reflect.StructField{
			Name: fmt.Sprintf("Field%03d", i),
			Type: fieldKindToType[serialization.FieldKind(f.Type)],
			Tag:  reflect.StructTag(fmt.Sprintf("json:\"%s\"", f.Name)),
		}
	}
	return reflect.StructOf(fs)
}

func (cm GenericCompact) Type() reflect.Type {
	return cm.ValueType
}

func (cm GenericCompact) TypeName() string {
	return cm.ValueTypeName
}

func (cm GenericCompact) Read(reader serialization.CompactReader) interface{} {
	rs := cm.readers
	v := reflect.New(cm.ValueType)
	for i, f := range cm.Fields {
		value := reflect.ValueOf(rs[i](reader, f.Name))
		if value.Interface() == nil {
			continue
		}
		v.Elem().Field(i).Set(value)
	}
	return v.Interface()
}

func (cm GenericCompact) Write(writer serialization.CompactWriter, value interface{}) {
	// TODO: implement me
	panic("implement me")
}

var fieldKindToType = map[serialization.FieldKind]reflect.Type{
	serialization.FieldKindNotAvailable: nil,
	serialization.FieldKindString:       reflect.TypeOf(""),
	serialization.FieldKindInt32:        reflect.TypeOf(int32(0)),
	//FieldKindBoolean        FieldKind = 1
	//FieldKindArrayOfBoolean FieldKind = 2
	//FieldKindInt8           FieldKind = 3
	//FieldKindArrayOfInt8    FieldKind = 4
	//FieldKindInt16                        FieldKind = 7
	//FieldKindArrayOfInt16                 FieldKind = 8
	//FieldKindInt32                        FieldKind = 9
	//FieldKindArrayOfInt32                 FieldKind = 10
	//FieldKindInt64                        FieldKind = 11
	//FieldKindArrayOfInt64                 FieldKind = 12
	//FieldKindFloat32                      FieldKind = 13
	//FieldKindArrayOfFloat32               FieldKind = 14
	//FieldKindFloat64                      FieldKind = 15
	//FieldKindArrayOfFloat64               FieldKind = 16
	//FieldKindString                       FieldKind = 17
	//FieldKindArrayOfString                FieldKind = 18
	//FieldKindDecimal                      FieldKind = 19
	//FieldKindArrayOfDecimal               FieldKind = 20
	//FieldKindTime                         FieldKind = 21
	//FieldKindArrayOfTime                  FieldKind = 22
	//FieldKindDate                         FieldKind = 23
	//FieldKindArrayOfDate                  FieldKind = 24
	//FieldKindTimestamp                    FieldKind = 25
	//FieldKindArrayOfTimestamp             FieldKind = 26
	//FieldKindTimestampWithTimezone        FieldKind = 27
	//FieldKindArrayOfTimestampWithTimezone FieldKind = 28
	//FieldKindCompact                      FieldKind = 29
	//FieldKindArrayOfCompact               FieldKind = 30
	//FieldKindNullableBoolean        FieldKind = 33
	//FieldKindArrayOfNullableBoolean FieldKind = 34
	//FieldKindNullableInt8           FieldKind = 35
	//FieldKindArrayOfNullableInt8    FieldKind = 36
	//FieldKindNullableInt16          FieldKind = 37
	//FieldKindArrayOfNullableInt16   FieldKind = 38
	//FieldKindNullableInt32          FieldKind = 39
	//FieldKindArrayOfNullableInt32   FieldKind = 40
	//FieldKindNullableInt64          FieldKind = 41
	//FieldKindArrayOfNullableInt64   FieldKind = 42
	//FieldKindNullableFloat32        FieldKind = 43
	//FieldKindArrayOfNullableFloat32 FieldKind = 44
	//FieldKindNullableFloat64        FieldKind = 45
	//FieldKindArrayOfNullableFloat64 FieldKind = 46

}
