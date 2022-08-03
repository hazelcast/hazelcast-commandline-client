package tuiutil

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type TestType struct {
	Test Color
}

func TestColor(t *testing.T) {
	tcs := []struct {
		name   string
		yml    string
		isErr  bool
		result lipgloss.Color
	}{
		{
			name: "valid color value",
			yml: `colorconfig:
    test: "#FFFFFF"`,
			result: lipgloss.Color("#FFFFFF"),
		},
		{
			name: "invalid color value",
			yml: `colorconfig:
    test: "#FF"`,
			isErr: true,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			v := struct {
				ColorConfig TestType
			}{}
			err := yaml.Unmarshal([]byte(tc.yml), &v)
			if tc.isErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.result, v.ColorConfig.Test.TerminalColor)
		})
	}
}
