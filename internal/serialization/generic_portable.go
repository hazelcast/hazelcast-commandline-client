package serialization

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type portableFieldReader func(r serialization.PortableReader, field string) any

type portableFieldWriter func(w serialization.PortableWriter, field string, value any)

type GenericPortable struct {
	Fields  []PortableField
	Name    string
	FID     int32
	CID     int32
	readers []portableFieldReader
	writers []portableFieldWriter
}

func NewGenericPortable(value GenericPortable) (*GenericPortable, error) {
	rs := make([]portableFieldReader, len(value.Fields))
	ws := make([]portableFieldWriter, len(value.Fields))
	for i, f := range value.Fields {
		pt := serialization.FieldDefinitionType(f.Type - 1)
		if pt < serialization.TypePortable {
			return nil, fmt.Errorf("invalid portable type: %d", f.Type)
		}
		r, ok := portableReaders[pt]
		if !ok {
			return nil, fmt.Errorf("reader not found for portable type: %d", f.Type)
		}
		rs[i] = r
		// writing is disabled for now --YT
	}
	return &GenericPortable{
		Fields:  value.Fields,
		Name:    value.Name,
		FID:     value.FID,
		CID:     value.CID,
		readers: rs,
		writers: ws,
	}, nil
}

func (g *GenericPortable) Clone() *GenericPortable {
	fs := make([]PortableField, len(g.Fields))
	copy(fs, g.Fields)
	return &GenericPortable{
		Fields:  fs,
		Name:    g.Name,
		FID:     g.FID,
		CID:     g.CID,
		readers: g.readers,
		writers: g.writers,
	}
}

func (g *GenericPortable) FactoryID() int32 {
	return g.FID
}

func (g *GenericPortable) ClassID() int32 {
	return g.CID
}

func (g *GenericPortable) WritePortable(pw serialization.PortableWriter) {
	ws := g.writers
	for i, v := range g.Fields {
		ws[i](pw, v.Name, v.Value)
	}
}

func (g *GenericPortable) ReadPortable(r serialization.PortableReader) {
	rs := g.readers
	for i, v := range g.Fields {
		g.Fields[i].Value = rs[i](r, v.Name)
	}
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

type GenericPortableFactory struct {
	classes   map[int32]*GenericPortable
	factoryID int32
}

func NewGenericPortableFactory(factoryID int32, items ...*GenericPortable) (*GenericPortableFactory, error) {
	cs := make(map[int32]*GenericPortable, len(items))
	for _, item := range items {
		if item.FID != factoryID {
			return nil, fmt.Errorf("serializer factoryID does not match factory ID")
		}
		cs[item.CID] = item
	}
	return &GenericPortableFactory{
		classes:   cs,
		factoryID: factoryID,
	}, nil
}

func (g GenericPortableFactory) Create(classID int32) serialization.Portable {
	cls, ok := g.classes[classID]
	if !ok {
		panic(fmt.Errorf("portable type for classID %d not found", classID))
	}
	return cls.Clone()
}

func (g GenericPortableFactory) FactoryID() int32 {
	return g.factoryID
}
