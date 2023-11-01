package str

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// SplitByComma splits a string by commas, and optionally removes empty items.
func SplitByComma(str string, removeEmpty bool) []string {
	var idx int
	if str == "" {
		return nil
	}
	ls := strings.Split(str, ",")
	for _, s := range ls {
		s = strings.TrimSpace(s)
		if s != "" || !removeEmpty {
			ls[idx] = s
			idx++
		}
	}
	return ls[0:idx:idx]

}

func ParseKeyValue(kvStr string) (string, string) {
	idx := strings.Index(kvStr, "=")
	if idx < 0 {
		return "", ""
	}
	return strings.TrimSpace(kvStr[:idx]), strings.TrimSpace(kvStr[idx+1:])
}

func MaybeShorten(s string, l int) string {
	if len(s) < 3 || len(s) < l {
		return s
	}
	return fmt.Sprintf("%s...", s[:l])
}

// SpacePaddedIntFormat returns the fmt string that can fit the given integer.
// The padding uses spaces.
func SpacePaddedIntFormat(maxValue int) string {
	if maxValue < 0 {
		panic("SpacePaddedIntFormat: cannot be negative")
	}
	return fmt.Sprintf("%%%dd", len(strconv.Itoa(maxValue)))
}

func Colorize(text string) string {
	if strings.HasPrefix(text, "OK ") {
		return fmt.Sprintf("    %s %s", color.GreenString("OK"), text[3:])
	}
	if strings.HasPrefix(text, "ERROR ") {
		return fmt.Sprintf(" %s %s", color.RedString("ERROR"), text[6:])
	}
	return text
}

func BytesToMegabytes(bytesStr string) (string, error) {
	bytes, err := strconv.ParseFloat(bytesStr, 64)
	if err != nil {
		return "", err
	}
	mb := bytes / (1024.0 * 1024.0)
	return fmt.Sprintf("%.2f MBs", mb), nil
}

func MsToSecs(ms string) (string, error) {
	milliseconds, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return "", err
	}
	seconds := float64(milliseconds) / 1000.0
	secondsStr := fmt.Sprintf("%.1f sec", seconds)
	return secondsStr, nil
}
