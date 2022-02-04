package types

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

var message = "The support for %s hasn't been implemented yet.\nIf you would like us to implement it, please attend a poll by at:\n%v and give a feedback.\nWe're happy to implement it quickly based on demand!\n"
var IssueURL = "https://docs.google.com/forms/d/e/1FAIpQLSfraK3Mu0u96rAau_OO-heoHE_z7gCZwX8Dw034RlzYf27M1Q/viewform?usp=sf_link"

type ColoredString struct {
	Word  string
	Color *color.Color
}

func (p ColoredString) String() string {
	if p.Word == "" {
		log.Panicln("word cannot be empty")
	}
	return p.Color.Sprint(p.Word)
}

func DecorateStringColorFgCyanWithUnderline(str string) *ColoredString {
	if str == "" {
		log.Panicln("given string cannot be empty")
	}
	return &ColoredString{
		Word:  str,
		Color: color.New(color.FgCyan).Add(color.Underline),
	}
}

type DsFakeDoorMessage struct {
	Name ColoredString
	Link ColoredString
}

func FakeDoorMessage(m DsFakeDoorMessage) string {
	return fmt.Sprintf(message, m.Name, m.Link)
}
