package table

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	console "github.com/nathan-fiscaletti/consolesize-go"
)

var consoleSize = console.GetConsoleSize

const (
	alignLeft = iota
	alignCenter
)

type TabularWriter struct {
	out         io.Writer
	rowsWritten int
	rowCount    int
	columnCount int
}

func NewTableWriter(out io.Writer) *TabularWriter {
	return &TabularWriter{out: out}
}

func (t *TabularWriter) Write(cells ...interface{}) error {
	t.initSize(cells)
	colWidth := (t.columnCount - (len(cells)+1)*2) / len(cells)
	line := makeTabularLine(colWidth, alignLeft, cells...)
	if _, err := fmt.Fprintln(t.out, line); err != nil {
		return err
	}
	t.rowsWritten++
	if t.rowsWritten == t.rowCount {
		// reset state for a new page
		t.rowsWritten = 0
	}
	return nil
}

func (t *TabularWriter) initSize(cells []interface{}) {
	if t.rowsWritten != 0 {
		// not start state
		return
	}
	// start state
	t.columnCount, t.rowCount = consoleSize()
	// minimum space required for => | ... | ... | ... |
	if t.columnCount < (len(cells)+1)*2+len(cells)*3 {
		t.columnCount = (len(cells)+1)*2 + len(cells)*3
	}
	if t.rowCount < 4 {
		t.rowCount = 4
	}
}

/*
WriteHeader outputs header for the table with the form:
+--------------------------------------+
| vegetables | fruit      | rank       |
+--------------------------------------+
*/
func (t *TabularWriter) WriteHeader(cells ...interface{}) error {
	t.initSize(cells)
	// colWidth = (totalWidth - numOfSeparators) / numOfColumns
	colWidth := (t.columnCount - (len(cells)+1)*2) / len(cells)
	unusedCellCount := (t.columnCount - (len(cells)+1)*2) % len(cells)
	headerBorder := fmt.Sprintf("+%s+\n", strings.Repeat("-", t.columnCount-unusedCellCount-2))
	// write upper border
	if _, err := fmt.Fprintf(t.out, headerBorder); err != nil {
		return err
	}
	// write column names
	line := makeTabularLine(colWidth, alignCenter, cells...)
	if _, err := fmt.Fprintln(t.out, line); err != nil {
		return err
	}
	// write lower border
	_, err := fmt.Fprintf(t.out, headerBorder)
	return err
}

func makeTabularLine(cellWidth int, alignment int, cells ...interface{}) string {
	var b strings.Builder
	b.WriteString("| ")
	for _, c := range cells {
		s := fmt.Sprint(c)
		td := runewidth.Truncate(s, cellWidth, "...")
		switch alignment {
		case alignLeft:
			b.WriteString(fmt.Sprintf("%-"+strconv.Itoa(cellWidth)+"s |", td))
		case alignCenter:
			b.WriteString(fmt.Sprintf("%"+strconv.Itoa(cellWidth)+"s |", td+strings.Repeat(" ", int((cellWidth-len(td))/2))))
		}

	}
	return b.String()
}
