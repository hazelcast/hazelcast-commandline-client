package table

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
)

const (
	AlignRight = -1
	AlignLeft  = 1
)

type runeWidthFn func(string, int) string

type Column struct {
	Header string
	Align  int
}

type Row []Column

type Table struct {
	rowIdx int64
	cfg    Config
	width  []int
	rwf    []runeWidthFn
	sep    string
}

type lines []string

func New(cfg Config) *Table {
	cfg.updateWithDefaults()
	return &Table{cfg: cfg}
}

func (t *Table) WriteHeader(cs Row) {
	t.width = make([]int, len(cs))
	t.rwf = make([]runeWidthFn, len(cs))
	t.updateAlignment(cs)
	if t.cfg.HeaderSeperator != "" {
		t.sep = t.makeSeparator(cs)
	}
	if t.sep != "" {
		withColor(t.cfg.TitleColor, func() {
			t.printf("%s", t.sep)
		})
		t.printf("\n")
	}
	withColor(t.cfg.TitleColor, func() {
		row := make([]string, len(cs))
		for i, h := range cs {
			row[i] = h.Header
		}
		t.printRow(row)
	})
	t.printf("\n")
	if t.sep != "" {
		withColor(t.cfg.TitleColor, func() {
			t.printf("%s", t.sep)
		})
		t.printf("\n")
	}
}

func (t *Table) WriteRow(row []string) {
	idx := atomic.AddInt64(&t.rowIdx, 1) - 1
	withColor(t.cfg.RowColors[idx&1], func() {
		t.printRow(row)
	})
	t.printf("\n")
}

func (t *Table) End() {
	if t.sep != "" {
		withColor(t.cfg.TitleColor, func() {
			t.printf("%s", t.sep)
		})
	}
	t.printf("\n")
}

func (t *Table) updateAlignment(row Row) {
	for i, h := range row {
		t.rwf[i] = runewidth.FillRight
		t.width[i] = h.Align
		if h.Align < 0 {
			t.rwf[i] = runewidth.FillLeft
			t.width[i] = -h.Align
		}
	}
}

func (t *Table) makeSeparator(row Row) string {
	bf := &bytes.Buffer{}
	rs := make([]string, len(row))
	for i, h := range row {
		rs[i] = h.Header
	}
	size := t.wPrintRow(bf, rs)
	repeat := size / len(t.cfg.HeaderSeperator)
	return strings.Repeat(t.cfg.HeaderSeperator, repeat)
}

func (t *Table) printRow(row []string) int {
	return t.wPrintRow(t.cfg.Stdout, row)
}

func (t *Table) wPrintRow(wr io.Writer, row []string) int {
	var maxLines int
	rowLines := make([]lines, len(row))
	// first iterate over all rows, splitting them to lines
	for i, r := range row {
		ls := splitLines(r, t.width[i])
		if len(ls) > maxLines {
			maxLines = len(ls)
		}
		rowLines[i] = ls
	}
	var rowWidth int
	var n int
	cf1 := t.cfg.CellFormat[1]
	for li := 0; li < maxLines; li++ {
		//first cell's format
		f := t.cfg.CellFormat[0]
		for ci, ls := range rowLines {
			var v string
			if len(ls) > li {
				v = ls[li]
			}
			w := t.width[ci]
			n += printf(wr, f, t.rwf[ci](v, w))
			// cell format except the first cell
			f = cf1
		}
		if li < maxLines-1 {
			printf(wr, "\n")
		}
	}
	rowWidth += n
	return rowWidth
}

func (t *Table) printf(format string, args ...any) int {
	return printf(t.cfg.Stdout, format, args...)
}

func printf(w io.Writer, format string, args ...any) int {
	n, err := fmt.Fprintf(w, format, args...)
	if err != nil {
		return 0
	}
	return n
}

func withColor(c *color.Color, f func()) {
	c.Set()
	f()
	color.Unset()
}

func splitLines(text string, maxWidth int) lines {
	// just a precaution to not allocate the universe
	if maxWidth > 65535 {
		maxWidth = 65535
	}
	if maxWidth == 0 {
		panic("splitLines: maxWidth cannot be 0")
	}
	const cr = '\n'
	line := make([]rune, maxWidth)
	var ls lines
	var cursor int
	for _, r := range text {
		if r == cr {
			ls = append(ls, string(line[:cursor]))
			cursor = 0
			continue
		}
		if cursor >= maxWidth {
			ls = append(ls, string(line[:cursor]))
			cursor = 0
		}
		line[cursor] = r
		cursor++
	}
	// the last line
	if cursor > 0 {
		ls = append(ls, string(line[:cursor]))
	}
	return ls
}
