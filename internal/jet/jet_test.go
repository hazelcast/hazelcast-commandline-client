package jet

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatch(t *testing.T) {
	b100 := newByteArray(100)
	testCases := []struct {
		b       *bytes.Buffer
		bs      int
		batches [][]byte
	}{
		{
			b:       bytes.NewBuffer(b100),
			bs:      100,
			batches: [][]byte{b100},
		},
		{
			b:       bytes.NewBuffer(b100),
			bs:      50,
			batches: [][]byte{b100[:50], b100[50:]},
		},
		{
			b:       bytes.NewBuffer(b100),
			bs:      30,
			batches: [][]byte{b100[0:30], b100[30:60], b100[60:90], b100[90:]},
		},
		{
			b:       bytes.NewBuffer(nil),
			bs:      1,
			batches: [][]byte{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("size %d, %d batches", tc.b.Len(), tc.bs), func(t *testing.T) {
			bb := newBatch(tc.b, tc.bs)
			result := [][]byte{}
			for {
				b, _, err := bb.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				bc := make([]byte, len(b))
				copy(bc, b)
				result = append(result, bc)
			}
			assert.Equal(t, tc.batches, result)
		})
	}
}

func newByteArray(size int) []byte {
	r := make([]byte, size)
	for i := 0; i < size; i++ {
		r[i] = byte(i)
	}
	return r
}
