package prompt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocument_FindPreviousWordStart(t *testing.T) {
	testCases := []struct {
		doc         Document
		charsToLeft int
	}{
		{
			doc: Document{
				Text:           "",
				cursorPosition: 0,
			},
			charsToLeft: 0,
		},
		{
			doc: Document{
				Text:           "abc",
				cursorPosition: 0,
			},
			charsToLeft: 0,
		},
		{
			doc: Document{
				Text:           "abc",
				cursorPosition: 1,
			},
			charsToLeft: 1,
		},
		{
			doc: Document{
				Text:           "abc",
				cursorPosition: 2,
			},
			charsToLeft: 2,
		},
		{
			doc: Document{
				Text:           "abc ",
				cursorPosition: 3,
			},
			charsToLeft: 3,
		},
		{
			doc: Document{
				Text:           "abc  ",
				cursorPosition: 4,
			},
			charsToLeft: 4,
		},
		{
			doc: Document{
				Text:           "abc x",
				cursorPosition: 4,
			},
			charsToLeft: 4,
		},
		{
			doc: Document{
				Text:           "abc xw",
				cursorPosition: 5,
			},
			charsToLeft: 1,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			assert.Equal(t, tc.charsToLeft, tc.doc.FindPreviousWordStart())
		})
	}
}
