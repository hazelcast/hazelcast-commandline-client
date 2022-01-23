package table

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTabularWriter_Write(t *testing.T) {
	consoleSize = func() (int, int) {
		return 50, 3
	}
	exampleCells := []interface{}{"someValue", 5.0, time.Date(1994, 8, 30, 0, 0, 0, 0, time.UTC)}
	exampleLine := "|someValue      |5              |1994-08-30 0...|"
	buffer := bytes.NewBuffer(nil)
	reader, writer := bufio.NewReader(buffer), buffer
	w := NewTableWriter(writer)
	steps := []struct {
		desc               string
		cells              []interface{}
		changeTerminalSize func()
		expectedLine       string
		expectedErr        error
	}{
		{
			desc:         "first line",
			cells:        exampleCells,
			expectedLine: exampleLine,
		},
	}
	for _, s := range steps {
		if s.changeTerminalSize != nil {
			// change terminal size and continue
			s.changeTerminalSize()
			continue
		}
		err := w.Write(s.cells...)
		assert.Equal(t, s.expectedErr, err)
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, s.expectedLine+"\n", line)
	}
}

func TestTabularWriter_WriteHeader(t *testing.T) {
	consoleSize = func() (int, int) {
		return 50, 3
	}
	buffer := bytes.NewBuffer(nil)
	w := NewTableWriter(buffer)
	if err := w.WriteHeader("vegetables", "fruit", "rank"); err != nil {
		t.Fatal(err)
	}
	expected := "+------------------------------------------------+\n" +
		"|   vegetables   |     fruit     |     rank      |\n" +
		"+------------------------------------------------+\n"
	assert.Equal(t, expected, buffer.String())
}
