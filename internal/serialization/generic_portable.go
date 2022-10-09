package serialization

import (
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type portableFieldReader func(r serialization.PortableReader, field string) any

type portableFieldWriter func(w serialization.PortableWriter, field string, value any)

type GenericPortable struct {
	Fields    []Field
	Name      string
	FactoryID int32
	ClassID   int32
}

type GenericPortableSerializer struct {
	gp      GenericPortable
	readers []portableFieldReader
	writers []portableFieldWriter
}

func NewGenericPortableSerializer(value GenericPortable) (*GenericPortableSerializer, error) {
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
		w, ok := portableWriters[pt]
		if !ok {
			return nil, fmt.Errorf("writer not found for portable type: %d", f.Type)
		}
		ws[i] = w
	}
	return &GenericPortableSerializer{
		gp:      value,
		readers: rs,
		writers: ws,
	}, nil
}

func (g *GenericPortableSerializer) Fields() []Field {
	return g.gp.Fields
}

func (g *GenericPortableSerializer) FactoryID() int32 {
	return g.gp.FactoryID
}

func (g *GenericPortableSerializer) ClassID() int32 {
	return g.gp.ClassID
}

func (g *GenericPortableSerializer) WritePortable(pw serialization.PortableWriter) {
	ws := g.writers
	for i, v := range g.gp.Fields {
		ws[i](pw, v.Name, v.Value)
	}
}

func (g *GenericPortableSerializer) ReadPortable(r serialization.PortableReader) {
	rs := g.readers
	for i, v := range g.gp.Fields {
		g.gp.Fields[i].Value = rs[i](r, v.Name)
	}
}

func (g *GenericPortableSerializer) String() string {
	fs := g.Fields()
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

type GenericPortableFactory struct {
	classes   map[int32]*GenericPortableSerializer
	factoryID int32
}

func NewGenericPortableFactory(factoryID int32, items ...*GenericPortableSerializer) (*GenericPortableFactory, error) {
	cs := make(map[int32]*GenericPortableSerializer, len(items))
	for _, item := range items {
		if item.gp.FactoryID != factoryID {
			return nil, fmt.Errorf("serializer factoryID does not match factory ID")
		}
		cs[item.gp.ClassID] = item
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
	return cls
}

func (g GenericPortableFactory) FactoryID() int32 {
	return g.factoryID
}
