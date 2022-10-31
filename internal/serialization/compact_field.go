package serialization

import (
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type CompactField struct {
	Name  string
	Type  CompactFieldType
	Value any
}

var compactReaders = map[serialization.FieldKind]compactFieldReader{
	serialization.FieldKindString: func(r serialization.CompactReader, field string) any {
		v := r.ReadString(field)
		if v == nil {
			return nil
		}
		return *v
	},
	serialization.FieldKindInt32: func(r serialization.CompactReader, field string) any {
		return r.ReadInt32(field)
	},
}

// writing compact values is not supported yet --YT
var compactWriters = map[serialization.FieldKind]compactFieldWriter{}

type CompactFieldType serialization.FieldKind

func (t *CompactFieldType) UnmarshalText(b []byte) error {
	s := strings.ToLower(string(b))
	switch s {
	case "":
		*t = CompactFieldType(serialization.FieldKindNotAvailable)
	case "int32":
		*t = CompactFieldType(serialization.FieldKindInt32)
	case "string":
		*t = CompactFieldType(serialization.FieldKindString)
	default:
		return fmt.Errorf("unknown type: %d", s)
	}
	return nil
}
