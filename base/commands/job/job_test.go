package job

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDToString(t *testing.T) {
	testCases := []struct {
		id int64
		s  string
	}{
		{
			id: 0,
			s:  "0000-0000-0000-0000",
		},
		{
			id: 665661962356523009,
			s:  "093c-e807-26c0-0001",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(strconv.FormatInt(tc.id, 10), func(t *testing.T) {
			assert.Equal(t, tc.s, idToString(tc.id))
		})
	}
}

func TestStringToID(t *testing.T) {
	testCases := []struct {
		s      string
		id     int64
		hasErr bool
	}{
		{
			s:  "0000-0000-0000-0000",
			id: 0,
		},
		{
			s:  "093c-e807-26c0-0001",
			id: 665661962356523009,
		},
		{
			s:  "665657305270124545",
			id: 665657305270124545,
		},
		{
			s:      "",
			hasErr: true,
		},
		{
			s:      "qqq",
			hasErr: true,
		},
		{
			s:      "---",
			hasErr: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(strconv.FormatInt(tc.id, 10), func(t *testing.T) {
			id, err := stringToID(tc.s)
			if err != nil {
				if tc.hasErr {
					return
				}
				t.Fatal(err)
			}
			assert.Equal(t, tc.id, id)
		})
	}
}

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
