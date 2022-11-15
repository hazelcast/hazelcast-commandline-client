package table

import (
	"fmt"
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

type Table struct {
	cfg    Config
	rowIdx int64
	width  []int
	rwf    []runeWidthFn
}

func New(cfg Config) *Table {
	cfg.updateWithDefaults()
	return &Table{cfg: cfg}
}

func (t *Table) WriteHeader(cs []Column) {
	t.width = make([]int, len(cs))
	t.rwf = make([]runeWidthFn, len(cs))
	t.updateAlignment(cs)
	var size int
	withColor(t.cfg.TitleColor, func() {
		row := make([]string, len(cs))
		for i, h := range cs {
			row[i] = h.Header
		}
		size = t.printRow(row)
	})
	t.printf("\n")
	if t.cfg.HeaderSeperator != "" {
		// TODO: support separators with more than one rune.
		repeat := size / len(t.cfg.HeaderSeperator)
		withColor(t.cfg.TitleColor, func() {
			t.printf("%s", strings.Repeat(t.cfg.HeaderSeperator, repeat))
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

func (t *Table) updateAlignment(cs []Column) {
	for i, h := range cs {
		t.rwf[i] = runewidth.FillRight
		t.width[i] = h.Align
		if h.Align < 0 {
			t.rwf[i] = runewidth.FillLeft
			t.width[i] = -h.Align
		}
	}
}

func (t *Table) printRow(row []string) int {
	var n int
	for i, v := range row {
		w := t.width[i]
		v = runewidth.Truncate(v, w, "")
		f := t.cfg.CellFormat[1]
		if i == 0 {
			f = t.cfg.CellFormat[0]
		}
		n += t.printf(f, t.rwf[i](v, w))
	}
	return n
}

func (t *Table) printf(format string, args ...any) int {
	// ignoring the error here
	n, err := fmt.Fprintf(t.cfg.Stdout, format, args...)
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
