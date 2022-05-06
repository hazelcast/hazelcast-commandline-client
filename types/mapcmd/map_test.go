package mapcmd

import (
	"bytes"
	"testing"
)

func TestObtainOrderingOfValues(t *testing.T) {
	for _, tc := range []struct {
		info string
		want []byte
		args []string
	}{
		{"value-short then file", []byte{'s', 'f'}, []string{"-v", "--value-file"}},
		{"file then value-short", []byte{'f', 's'}, []string{"--value-file", "-v"}},
		{"file-short then value", []byte{'f', 's'}, []string{"-f", "--value"}},
		{"value then file-short", []byte{'s', 'f'}, []string{"--value", "-f"}},
		{"value-short then file-short", []byte{'s', 'f'}, []string{"-v", "-f"}},
		{"file-short then value-short", []byte{'f', 's'}, []string{"-f", "-v"}},
		{"empty", nil, []string{}},
	} {
		t.Run(tc.info, func(t *testing.T) {
			gotvOrder, _ := ObtainOrderingOfValueFlags(tc.args)
			if !bytes.Equal(tc.want, gotvOrder) {
				t.Errorf("want %v got %v", tc.want, gotvOrder)
			}
		})
	}
}
