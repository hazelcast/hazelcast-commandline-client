package viewer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	HiddenTmpDirectoryName = ".termdbms"
	SQLSnippetsFile        = "snippets.termdbms"
)

func TruncateIfApplicable(m *TuiModel, conv string) (s string) {
	max := 0
	viewportWidth := m.Viewport.Width
	cellWidth := m.CellWidth()
	if m.UI.RenderSelection || m.UI.ExpandColumn > -1 {
		max = viewportWidth
	} else {
		max = cellWidth
	}
	if strings.Count(conv, "\n") > 0 {
		conv = SplitLines(conv)[0]
	}
	textWidth := lipgloss.Width(conv)
	minVal := Min(textWidth, max)

	if max == minVal && textWidth >= max { // truncate
		s = conv[:minVal]
		s = s[:lipgloss.Width(s)-3] + "..."
	} else {
		s = conv
	}
	return s
}

func GetStringRepresentationOfInterface(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	} else if i, ok := val.(int64); ok { // these default to int64 so not sure how this would affect 32 bit systems TODO
		return fmt.Sprintf("%d", i)
	} else if i, ok := val.(int32); ok { // these default to int32 so not sure how this would affect 32 bit systems TODO
		return fmt.Sprintf("%d", i)
	} else if i, ok := val.(float64); ok {
		return fmt.Sprintf("%.2f", i)
	} else if i, ok := val.(float32); ok {
		return fmt.Sprintf("%.2f", i)
	} else if t, ok := val.(time.Time); ok {
		str := t.String()
		return str
	} else if val == nil {
		return "NULL"
	}
	return ""
}

func SplitLines(s string) []string {
	var lines []string
	if strings.Count(s, "\n") == 0 {
		return append(lines, s)
	}
	reader := strings.NewReader(s)
	sc := bufio.NewScanner(reader)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

func GetScrollDownMaximumForSelection(m *TuiModel) int {
	max := 0
	if m.UI.RenderSelection {
		conv, _ := FormatJson(m.Data().EditTextBuffer)
		lines := SplitLines(conv)
		max = len(lines)
	} else if m.UI.FormatModeEnabled {
		max = len(SplitLines(DisplayFormatText(m)))
	} else {
		return len(m.GetColumnData())
	}
	return max
}

// FormatJson is some more code I stole off stackoverflow
func FormatJson(str string) (string, error) {
	b := []byte(str)
	if !json.Valid(b) { // return original string if not json
		return str, errors.New("this is not valid JSON")
	}
	var formattedJson bytes.Buffer
	if err := json.Indent(&formattedJson, b, "", "    "); err != nil {
		return "", err
	}
	return formattedJson.String(), nil
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
