package serialization

import (
	"sort"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type portableFieldReader func(r serialization.PortableReader, field string) Column

type GenericPortable struct {
	Fields ColumnMap
	FID    int32
	CID    int32
}

func (p *GenericPortable) FactoryID() int32 {
	return p.FID
}

func (p *GenericPortable) ClassID() int32 {
	return p.CID
}

func (p *GenericPortable) WritePortable(w serialization.PortableWriter) {
	panic("serialization.GenericPortable.WritePortable is not supposed to be called")
}

func (g *GenericPortable) ReadPortable(r serialization.PortableReader) {
	panic("serialization.GenericPortable.ReadPortable is not supposed to be called")
}

func (g *GenericPortable) Text() string {
	return g.Fields.Text()
}

func (g *GenericPortable) JSONValue() (any, error) {
	return g.Fields.JSONValue()
}

type GenericPortableSerializer struct{}

func NewGenericPortableSerializer() *GenericPortableSerializer {
	return &GenericPortableSerializer{}
}

func (gs GenericPortableSerializer) CreatePortableValue(factoryID, classID int32) serialization.Portable {
	return &GenericPortable{
		FID: factoryID,
		CID: classID,
	}
}

func (gs GenericPortableSerializer) ReadPortableWithClassDefinition(portable serialization.Portable, cd *serialization.ClassDefinition, reader serialization.PortableReader) {
	v := portable.(*GenericPortable)
	for name, f := range cd.Fields {
		v.Fields = append(v.Fields, portableReaders[f.Type](reader, name))
	}
	// sort fields
	sort.Slice(v.Fields, func(i, j int) bool {
		return v.Fields[i].Name < v.Fields[j].Name
	})
}
