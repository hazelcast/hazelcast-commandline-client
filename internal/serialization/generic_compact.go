package serialization

import (
	"reflect"
	"sort"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type CompactFieldType serialization.FieldKind

type CompactField struct {
	Name  string
	Type  CompactFieldType
	Value any
}

type compactFieldReader func(r serialization.CompactReader, field string) Column

type SchemaInfo struct {
	Type     reflect.Type
	TypeName string
}

type GenericCompactDeserializer struct{}

func (cm GenericCompactDeserializer) Read(schema *hazelcast.Schema, reader serialization.CompactReader) interface{} {
	fds := schema.FieldDefinitions()
	cs := make(ColumnMap, len(fds))
	for i, fd := range fds {
		cs[i] = compactReaders[fd.Kind](reader, fd.Name)
	}
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].Name < cs[j].Name
	})
	return cs
}
