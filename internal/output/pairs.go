package output

import (
	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func DecodePairs(ic *hazelcast.ClientInternal, pairs []hazelcast.Pair, showType bool) []Row {
	rows := make([]Row, 0, len(pairs))
	vs := []any{}
	for _, pair := range pairs {
		row := make(Row, 0, 4)
		kt, key := ensureTypeValue(ic, pair.Key.(hazelcast.Data))
		row = append(row, NewKeyColumn(kt, key))
		if showType {
			row = append(row, NewKeyTypeColumn(kt))
		}
		vt, value := ensureTypeValue(ic, pair.Value.(hazelcast.Data))
		vs = append(vs, value)
		row = append(row, NewValueColumn(vt, value))
		if showType {
			row = append(row, NewValueTypeColumn(vt))
		}
		rows = append(rows, row)
	}
	return rows
}

func ensureTypeValue(ic *hazelcast.ClientInternal, data hazelcast.Data) (int32, any) {
	t := data.Type()
	v, err := ic.DecodeData(data)
	if err != nil {
		v = serialization.NondecodedType(serialization.TypeToLabel(t))
	}
	return t, v
}
