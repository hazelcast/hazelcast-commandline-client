package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/errors"
	iserialization "github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type JSONResult struct {
	rp RowProducer
}

func NewJSONResult(rp RowProducer) *JSONResult {
	return &JSONResult{rp: rp}
}

func (jr *JSONResult) Serialize(ctx context.Context, w io.Writer) (int, error) {
	var n int
	cr := []byte{'\n'}
	for {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		row, ok := jr.rp.NextRow(ctx)
		if !ok {
			return n, nil
		}
		m := make(map[string]any, len(row))
		for _, col := range row {
			v, err := columnToJSONValue(col)
			if err != nil {
				continue
			}
			m[col.Name] = v
		}
		b, err := json.Marshal(m)
		if err != nil {
			return 0, fmt.Errorf("json marshalling result: %w", err)
		}
		wn, err := w.Write(b)
		if err != nil {
			return 0, fmt.Errorf("serializing result: %w", err)
		}
		wn, err = w.Write(cr)
		if err != nil {
			return 0, fmt.Errorf("serializing result: %w", err)
		}
		n += wn
	}
}

func columnToJSONValue(col Column) (any, error) {
	switch col.Type {
	case iserialization.TypeNil:
		return nil, nil
	case iserialization.TypePortable, iserialization.TypeCompact:
		return col.Value, nil
	case iserialization.TypeDataSerializable:
		return nil, errors.ErrNotDecoded
	case iserialization.TypeByte, iserialization.TypeBool, iserialization.TypeUInt16,
		iserialization.TypeInt16, iserialization.TypeInt32, iserialization.TypeInt64,
		iserialization.TypeFloat32, iserialization.TypeFloat64, iserialization.TypeString,
		iserialization.TypeByteArray, iserialization.TypeBoolArray, iserialization.TypeUInt16Array,
		iserialization.TypeInt16Array, iserialization.TypeInt32Array, iserialization.TypeInt64Array,
		iserialization.TypeFloat32Array, iserialization.TypeFloat64Array, iserialization.TypeStringArray:
		return col.Value, nil
	case iserialization.TypeUUID:
		return col.Value.(types.UUID).String(), nil
	case iserialization.TypeSimpleEntry, iserialization.TypeSimpleImmutableEntry:
		return nil, errors.ErrNotDecoded
	case iserialization.TypeJavaClass:
		return col.Value.(string), nil
	case iserialization.TypeJavaDate:
		return col.Value.(time.Time).Format(time.RFC3339), nil
	case iserialization.TypeJavaBigInteger:
		return col.Value.(*big.Int).String(), nil
	case iserialization.TypeJavaDecimal:
		return iserialization.MarshalDecimal(col.Value), nil
	case iserialization.TypeJavaArray, iserialization.TypeJavaArrayList, iserialization.TypeJavaLinkedList:
		return col.Value, nil
	case iserialization.TypeJavaLocalDate:
		sr, err := iserialization.MarshalLocalDate(col.Value)
		if err != nil {
			return nil, errors.ErrNotDecoded
		} else if sr == nil {
			return nil, nil
		} else {
			return *sr, nil
		}
	case iserialization.TypeJavaLocalTime:
		sr, err := iserialization.MarshalLocalTime(col.Value)
		if err != nil {
			return nil, errors.ErrNotDecoded
		} else if sr == nil {
			return nil, nil
		} else {
			return *sr, nil
		}
	case iserialization.TypeJavaLocalDateTime:
		sr, err := iserialization.MarshalLocalDateTime(col.Value)
		if err != nil {
			return nil, errors.ErrNotDecoded
		} else if sr == nil {
			return nil, nil
		} else {
			return *sr, nil
		}
	case iserialization.TypeJavaOffsetDateTime:
		sr, err := iserialization.MarshalOffsetDateTime(col.Value)
		if err != nil {
			return nil, errors.ErrNotDecoded
		} else if sr == nil {
			return nil, nil
		} else {
			return *sr, nil
		}
	case iserialization.TypeJSONSerialization:
		return col.Value, nil
	}
	return nil, errors.ErrNotDecoded
}
