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
	width       int
}

func NewTableWriter(out io.Writer) *TabularWriter {
	return &TabularWriter{out: out}
}

func (t *TabularWriter) Write(cells ...interface{}) error {
	t.initSize(cells)
	// colWidth = (totalWidth - numOfSeparators) / numOfColumns
	colWidth := (t.width - len(cells) - 1) / len(cells)
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
	t.width, t.rowCount = consoleSize()
	// minimum space required for => | ... | ... | ... |
	if t.width < (len(cells)+1)*2+len(cells)*3 {
		t.width = (len(cells)+1)*2 + len(cells)*3
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
const cornersWithPlusSign = 2

func (t *TabularWriter) WriteHeader(cells ...interface{}) error {
	t.initSize(cells)
	// colWidth = (totalWidth - numOfSeparators) / numOfColumns
	colWidth := (t.width - (len(cells) + 1)) / len(cells)
	effectiveLineWidth := colWidth*len(cells) + len(cells) + 1 - cornersWithPlusSign
	headerBorder := fmt.Sprintf("+%s+\n", strings.Repeat("-", effectiveLineWidth))
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

const paddingFromSeparators = 2

func makeTabularLine(cellWidth, alignment int, cells ...interface{}) string {
	var b strings.Builder
	for _, c := range cells {
		s := fmt.Sprint(c)
		td := fmt.Sprintf(" %s ", runewidth.Truncate(s, cellWidth-paddingFromSeparators, "..."))
		switch alignment {
		case alignLeft:
			b.WriteString(fmt.Sprintf("|%-"+strconv.Itoa(cellWidth)+"s", td))
		case alignCenter:
			b.WriteString(fmt.Sprintf("|%"+strconv.Itoa(cellWidth)+"s", td+strings.Repeat(" ", (cellWidth-len(td))/2)))
		}
	}
	b.WriteString("|")
	return b.String()
}
