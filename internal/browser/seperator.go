package browser

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// progressState is the global flag to show/hide progress
var progressState int32

const (
	HideProgress = iota
	ShowProgress
)

var spinnerWidget = spinner.New()

func changeProgress(ps int32) {
	switch ps {
	case HideProgress:
		if atomic.CompareAndSwapInt32(&progressState, ShowProgress, HideProgress) {
			return
		}
	case ShowProgress:
		if atomic.CompareAndSwapInt32(&progressState, HideProgress, ShowProgress) {
			return
		}
	}
}

type SeparatorWithProgress struct {
	length int
}

func (s *SeparatorWithProgress) Init() tea.Cmd {
	return spinnerWidget.Tick
}

func (s *SeparatorWithProgress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		s.length = msg.Width
	}
	spinnerWidget, cmd = spinnerWidget.Update(msg)
	return s, cmd
}

func (s *SeparatorWithProgress) View() string {
	var baseMsg string
	if atomic.LoadInt32(&progressState) == ShowProgress {
		baseMsg = fmt.Sprintf(" %s Executing query ", spinnerWidget.View())
	}
	return strings.Repeat("─", max(0, s.length-len(baseMsg))) + baseMsg
}
