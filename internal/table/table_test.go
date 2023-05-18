package table

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitLines(t *testing.T) {
	testCases := []struct {
		name     string
		in       string
		out      lines
		maxWidth int
	}{
		{
			name: "empty",
			in:   "",
			out:  nil,
		},
		{
			name: "single line short",
			in:   "short-line",
			out:  lines{"short-line"},
		},
		{
			name: "multi line short",
			in:   "short1\nshort2",
			out:  lines{"short1", "short2"},
		},
		{
			name:     "single line long",
			in:       "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			out:      lines{"Lorem ipsum dolor sit ame", "t, consectetur adipiscing", " elit, sed do eiusmod tem", "por incididunt ut labore ", "et dolore magna aliqua."},
			maxWidth: 25,
		},
		{
			name:     "multi line long",
			in:       "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
			out:      lines{"Lorem ipsum dolor sit ame", "t, consectetur adipiscing", " elit, sed do eiusmod tem", "por incididunt ut labore ", "et dolore magna aliqua.", "Ut enim ad minim veniam, ", "quis nostrud exercitation", " ullamco laboris nisi ut ", "aliquip ex ea commodo con", "sequat."},
			maxWidth: 25,
		},
		{
			name: "single line unicode characters",
			in:   "ağaç ve şarkı, 木と歌",
			out:  lines{"ağaç ve şarkı, 木と歌"},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			w := tc.maxWidth
			if w == 0 {
				w = 20
			}
			ls := splitLines(tc.in, w)
			assert.Equal(t, tc.out, ls)
		})
	}
}

func TestTable(t *testing.T) {
	out := &bytes.Buffer{}
	cfg := Config{
		Stdout:     out,
		CellFormat: [2]string{" %s ", "| %s "},
	}
	cfg.HeaderSeperator = "-"
	tb := New(cfg)
	headers := []string{"Col 1", "Col 2"}
	hd := make(Row, len(headers))
	for i, h := range headers {
		hd[i] = Column{
			Header: h,
			Align:  len(h) * 2,
		}
	}
	tb.WriteHeader(hd)
	tb.WriteRow([]string{
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ",
		"col2 line1",
	})
	tb.End()
	target := `-------------------------
 Col 1      | Col 2      
-------------------------
 Lorem ipsu | col2 line1 
 m dolor si |            
 t amet, co |            
 nsectetur  |            
 adipiscing |            
  elit, sed |            
  do eiusmo |            
 d tempor i |            
 ncididunt  |            
 ut labore  |            
 et dolore  |            
 magna aliq |            
 ua.        |            
-------------------------
`
	assert.Equal(t, target, out.String())
}
