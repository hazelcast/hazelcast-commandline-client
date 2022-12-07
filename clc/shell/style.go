package shell

import (
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/styles"
)

func defaultStyle() *chroma.Style {
	return chroma.MustNewStyle("clc-default", chroma.StyleEntries{
		chroma.TextWhitespace:    "#888888",
		chroma.Background:        "#ffffff bg:#111111",
		chroma.GenericOutput:     "#444444 bg:#222222",
		chroma.Keyword:           "#fb660a bold",
		chroma.KeywordPseudo:     "nobold",
		chroma.LiteralNumber:     "#0086f7 bold",
		chroma.NameTag:           "#fb660a bold",
		chroma.NameVariable:      "#fb660a",
		chroma.Comment:           "italic",
		chroma.NameAttribute:     "#ff0086 bold",
		chroma.LiteralString:     "#0086d2",
		chroma.NameFunction:      "#ff0086 bold",
		chroma.GenericHeading:    "#ffffff bold",
		chroma.KeywordType:       "#cdcaa9 bold",
		chroma.GenericSubheading: "#ffffff bold",
		chroma.NameConstant:      "#0086d2",
		chroma.CommentPreproc:    "#ff0007 bold",
	})
}

func init() {
	styles.Register(defaultStyle())
}
