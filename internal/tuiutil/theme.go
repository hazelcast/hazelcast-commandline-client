package tuiutil

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	Ascii bool
	Faint bool
)

type Color struct {
	lipgloss.TerminalColor
}

var InvalidHexColorErr = fmt.Errorf(`color values should hex strings as in "#ffffff"`)

func NewColor(color lipgloss.TerminalColor) *Color {
	return &Color{color}
}

func (c *Color) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return InvalidHexColorErr
	}
	if _, err := colorful.Hex(s); err != nil {
		return InvalidHexColorErr
	}
	c.TerminalColor = lipgloss.Color(s)
	return nil
}

type ColorPalette struct {
	HeaderBackground *Color
	Border           *Color
	ResultText       *Color
	HeaderForeground *Color
	Highlight        *Color
	FooterForeground *Color
}

// styling functions
var (
	Highlight = func() lipgloss.TerminalColor {
		return GetTheme().Highlight
	} // change to whatever
	HeaderBackground = func() lipgloss.TerminalColor {
		return GetTheme().HeaderBackground
	}
	HeaderForeground = func() lipgloss.TerminalColor {
		return GetTheme().HeaderForeground
	}
	FooterForeground = func() lipgloss.TerminalColor {
		return GetTheme().FooterForeground
	}
	BorderColor = func() lipgloss.TerminalColor {
		return GetTheme().Border
	}
	TextColor = func() lipgloss.TerminalColor {
		return GetTheme().ResultText
	}
)

func GetTheme() *ColorPalette {
	return ThemesMap[SelectedTheme]
}

var noColor = lipgloss.NoColor{}

type Theme string

const (
	Default   = "default"
	NoColor   = "no-color"
	Nord      = "nord"
	Solarized = "solarized"
)

var themeNames = []string{
	"default",
	"no-color",
	"nord",
	"solarized",
}

func SetTheme(theme string) error {
	for _, t := range themeNames {
		if strings.ToLower(theme) == t {
			SelectedTheme = t
			return nil
		}
	}
	return fmt.Errorf("invalid theme name")
}

var (
	SelectedTheme = Default // adjust to the background
	ValidThemes   = []string{
		Default,
		NoColor,
		Nord,
		Solarized,
	}
	ThemesMap = map[string]*ColorPalette{
		Solarized: {
			HeaderBackground: NewColor(lipgloss.Color("#5e81ac")),
			Border:           NewColor(lipgloss.Color("#eceff4")),
			ResultText:       NewColor(lipgloss.Color("#eceff4")),
			HeaderForeground: NewColor(lipgloss.Color("#eceff4")),
			Highlight:        NewColor(lipgloss.Color("#88c0d0")),
			FooterForeground: NewColor(lipgloss.Color("#b48ead")),
		},
		Nord: {
			HeaderBackground: NewColor(lipgloss.Color("#383838")),
			Border:           NewColor(lipgloss.Color("#FFFFFF")),
			ResultText:       NewColor(lipgloss.Color("#FFFFFF")),
			HeaderForeground: NewColor(lipgloss.Color("#FFFFFF")),
			Highlight:        NewColor(lipgloss.Color("#A0A0A0")),
			FooterForeground: NewColor(lipgloss.Color("#C2C2C2")),
		},
		NoColor: { // no color
			HeaderBackground: NewColor(noColor),
			Border:           NewColor(noColor),
			ResultText:       NewColor(noColor),
			HeaderForeground: NewColor(noColor),
			Highlight:        NewColor(noColor),
			FooterForeground: NewColor(noColor),
		},
		Default: { // default
			HeaderBackground: NewColor(lipgloss.Color("#383838")),
			Border:           NewColor(noColor),
			ResultText:       NewColor(noColor),
			HeaderForeground: NewColor(lipgloss.Color("#FFFFFF")),
			Highlight:        NewColor(lipgloss.Color("#A0A0A0")),
			FooterForeground: NewColor(lipgloss.Color("#C2C2C2")),
		},
	}
)
