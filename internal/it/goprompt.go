package it

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	goprompt "github.com/hazelcast/hazelcast-commandline-client/internal/go-prompt"
)

// GoPromptOutputWriter acts as a writer impl. for go-prompt. It discards ansi escape sequences and
// mimics a few of the control sequences with just text.
type GoPromptOutputWriter struct {
	b     *bytes.Buffer
	flush chan string
}

func NewGoPromptOutputWriter(flushBufferCapacity int) GoPromptOutputWriter {
	var gpw GoPromptOutputWriter
	var b bytes.Buffer
	gpw.flush = make(chan string, flushBufferCapacity)
	gpw.b = &b
	return gpw
}

func (gpw GoPromptOutputWriter) ReadLatestFlushWithTimeout(timeout time.Duration) string {
	var result string
	var finish bool
	for timeout := time.After(timeout); !finish; {
		select {
		case <-timeout:
			finish = true
		case content := <-gpw.flush:
			fmt.Println(content)
			result = content
		}
	}
	return result
}

func (gpw GoPromptOutputWriter) WriteRaw(data []byte) {
	gpw.b.Write(data)
}

func (gpw GoPromptOutputWriter) Write(data []byte) {
	fmt.Println("bla")
}

func (gpw GoPromptOutputWriter) WriteRawStr(data string) {
	fmt.Println("bla")
}

func (gpw GoPromptOutputWriter) WriteStr(data string) {
	gpw.b.WriteString(data)
}

func (gpw GoPromptOutputWriter) Flush() error {
	if gpw.b.Len() == 0 {
		return nil
	}
	gpw.flush <- gpw.b.String()
	gpw.b.Reset()
	return nil
}

func (gpw GoPromptOutputWriter) EraseScreen() {}

func (gpw GoPromptOutputWriter) EraseUp() {}

func (gpw GoPromptOutputWriter) EraseDown() {}

func (gpw GoPromptOutputWriter) EraseStartOfLine() {}

func (gpw GoPromptOutputWriter) EraseEndOfLine() {}

func (gpw GoPromptOutputWriter) EraseLine() {}

func (gpw GoPromptOutputWriter) ShowCursor() {}

func (gpw GoPromptOutputWriter) HideCursor() {}

func (gpw GoPromptOutputWriter) CursorGoTo(row, col int) {}

func (gpw GoPromptOutputWriter) CursorUp(n int) {}

func (gpw GoPromptOutputWriter) CursorDown(n int) {
	// this used as new line
	for i := 0; i < n; i++ {
		gpw.b.WriteString("\n")
	}
}

func (gpw GoPromptOutputWriter) CursorForward(n int) {
	// maybe
}

func (gpw GoPromptOutputWriter) CursorBackward(n int) {
	// maybe
}

func (gpw GoPromptOutputWriter) AskForCPR() {}

func (gpw GoPromptOutputWriter) SaveCursor() {}

func (gpw GoPromptOutputWriter) UnSaveCursor() {}

func (gpw GoPromptOutputWriter) ScrollDown() {}

func (gpw GoPromptOutputWriter) ScrollUp() {}

func (gpw GoPromptOutputWriter) SetTitle(title string) {}

func (gpw GoPromptOutputWriter) ClearTitle() {}

func (gpw GoPromptOutputWriter) SetColor(fg, bg goprompt.Color, bold bool) {
	// do nothing to eliminate ansi sequences
}

func NewGoPromptInput() GoPromptInput {
	var gpi GoPromptInput
	// these default values should work for most cases
	gpi.input = make(chan []byte, 100)
	gpi.r, gpi.c = 1000, 2000
	return gpi
}

// GoPromptInput can be used to simulate user interaction with the prompt
type GoPromptInput struct {
	input chan []byte
	r, c  uint16
}

// WriteInput may block if the channel is full
func (gpi GoPromptInput) WriteInput(inp []byte) {
	gpi.input <- inp
}

func (gpi GoPromptInput) ChangeTerminalSize(rows, cols uint16) {
	gpi.r, gpi.c = rows, cols
}

func (gpi GoPromptInput) Setup() error {
	return nil
}

func (gpi GoPromptInput) TearDown() error {
	return nil
}

func (gpi GoPromptInput) GetWinSize() *goprompt.WinSize {
	return &goprompt.WinSize{
		Row: gpi.r,
		Col: gpi.c,
	}
}

func (gpi GoPromptInput) Read() ([]byte, error) {
	select {
	case b, ok := <-gpi.input:
		if ok {
			return b, nil
		}
		return nil, io.EOF
	default:
		// don't know why but go prompts works this way.
		return nil, errors.New("nothing to read")
	}
}
