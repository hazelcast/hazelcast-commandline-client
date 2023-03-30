package serialization

import (
	"encoding/json"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type portableFieldReader func(r serialization.PortableReader, field string) any

type portableFieldWriter func(w serialization.PortableWriter, field string, value any)

type GenericPortable struct {
	Fields []PortableField
	FID    int32
	CID    int32
}

func (p *GenericPortable) FactoryID() int32 {
	return p.FID
}

func (p *GenericPortable) ClassID() int32 {
	return p.CID
}

func (p *GenericPortable) WritePortable(writer serialization.PortableWriter) {
	panic("serialization.GenericPortable.WritePortable is not supposed to be called")
}

func (g *GenericPortable) ReadPortable(r serialization.PortableReader) {
	panic("serialization.GenericPortable.ReadPortable is not supposed to be called")
}

func (g *GenericPortable) String() string {
	fs := g.Fields
	if len(fs) == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.WriteString(fs[0].String())
	for _, f := range fs[1:] {
		// middle dot
		sb.WriteString("\u00b7")
		sb.WriteString(f.String())
	}
	return sb.String()
}

func (g *GenericPortable) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(g.Fields))
	for _, f := range g.Fields {
		m[f.Name] = g.marshalField(f)
	}
	return json.Marshal(m)
}

func (g *GenericPortable) marshalField(f PortableField) any {
	switch f.Type {
	case PortableTypeTime:
		// ignoring the error
		sr, _ := MarshalLocalTime(f.Value)
		if sr == nil {
			return nil
		}
		return *sr
	case PortableTypeDate:
		// ignoring the error
		sr, _ := MarshalLocalDateTime(f.Value)
		if sr == nil {
			return nil
		}
		return *sr
	case PortableTypeTimestamp:
		// ignoring the error
		sr, _ := MarshalLocalDateTime(f.Value)
		if sr == nil {
			return nil
		}
		return *sr
	case PortableTypeTimestampWithTimezone:
		// ignoring the error
		sr, _ := MarshalOffsetDateTime(f.Value)
		if sr == nil {
			return nil
		}
		return *sr
	}
	return f.Value
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
	for name, f := range cd.Fields {
		v := portable.(*GenericPortable)
		v.Fields = append(v.Fields, PortableField{
			Name:  f.Name,
			Type:  PortableFieldType(f.Type + 1),
			Value: portableReaders[f.Type](reader, name),
		})
	}
}
