package tuiutil

import (
	"context"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

const DefaultBlinkSpeed = time.Millisecond * 530

// Internal messages for clipboard operations.
type pasteMsg string
type pasteErrMsg struct{ error }

// EchoMode sets the input behavior of the text input field.
type EchoMode int

const (
	// EchoNormal displays text as is. This is the default behavior.
	EchoNormal EchoMode = iota

	// EchoPassword displays the EchoCharacter mask instead of actual
	// characters.  This is commonly used for password fields.
	EchoPassword

	// EchoNone displays nothing as characters are entered. This is commonly
	// seen for password fields on the command line.
	EchoNone

	// EchoOnEdit
)

// blinkCtx manages cursor blinking.
type blinkCtx struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// CursorMode describes the behavior of the cursor.
type CursorMode int

// Available cursor modes.
const (
	CursorBlink CursorMode = iota
	CursorStatic
	CursorHide
)

// String returns a the cursor mode in a human-readable format. This method is
// provisional and for informational purposes only.
func (c CursorMode) String() string {
	return [...]string{
		"blink",
		"static",
		"hidden",
	}[c]
}

// Paste is a command for pasting from the clipboard into the text input.
func Paste() tea.Msg {
	str, err := clipboard.ReadAll()
	if err != nil {
		return pasteErrMsg{err}
	}
	return pasteMsg(str)
}
