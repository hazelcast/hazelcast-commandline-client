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
	var n int
	for i, v := range row {
		w := t.width[i]
		v = runewidth.Truncate(v, w, "")
		f := t.cfg.CellFormat[1]
		if i == 0 {
			f = t.cfg.CellFormat[0]
		}
		n += printf(wr, f, t.rwf[i](v, w))
	}
	return n
}

func (t *Table) printf(format string, args ...any) int {
	return printf(t.cfg.Stdout, format, args...)
}

func printf(w io.Writer, format string, args ...any) int {
	// ignoring the error here
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
