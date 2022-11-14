package table

import (
	"io"
	"os"

	"github.com/fatih/color"
)

type Config struct {
	Stdout     io.Writer
	TitleColor *color.Color
	RowColors  [2]*color.Color
	RowFormat  string
}

func (c *Config) updateWithDefaults() {
	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}
	if c.TitleColor == nil {
		c.TitleColor = color.New(color.BgCyan, color.FgHiWhite, color.Bold)
	}
	if c.RowColors[0] == nil {
		c.RowColors[0] = color.New()
	}
	if c.RowColors[1] == nil {
		c.RowColors[1] = color.New(color.BgWhite, color.FgBlack)
	}
	if c.RowFormat == "" {
		c.RowFormat = " %s "
	}
}
