package table

import (
	"io"
	"os"

	"github.com/fatih/color"
)

type Config struct {
	Stdout     io.Writer
	TitleColor *color.Color
	// RowColors is the colors of odd and even rows
	RowColors [2]*color.Color
	// CellFormat is the cell format for the first and rest of the columns
	CellFormat      [2]string
	HeaderSeperator string
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
	if c.CellFormat[0] == "" {
		c.CellFormat[0] = " %s "
	}
	if c.CellFormat[1] == "" {
		c.CellFormat[1] = c.CellFormat[0]
	}
}
