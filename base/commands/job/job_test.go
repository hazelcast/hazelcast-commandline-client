package job

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDToString(t *testing.T) {
	testCases := []struct {
		id int64
		s  string
	}{
		{
			id: 0,
			s:  "0000-0000-0000-0000",
		},
		{
			id: 665661962356523009,
			s:  "093c-e807-26c0-0001",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(strconv.FormatInt(tc.id, 10), func(t *testing.T) {
			assert.Equal(t, tc.s, idToString(tc.id))
		})
	}
}

func TestStringToID(t *testing.T) {
	testCases := []struct {
		s      string
		id     int64
		hasErr bool
	}{
		{
			s:  "0000-0000-0000-0000",
			id: 0,
		},
		{
			s:  "093c-e807-26c0-0001",
			id: 665661962356523009,
		},
		{
			s:  "665657305270124545",
			id: 665657305270124545,
		},
		{
			s:      "",
			hasErr: true,
		},
		{
			s:      "qqq",
			hasErr: true,
		},
		{
			s:      "---",
			hasErr: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(strconv.FormatInt(tc.id, 10), func(t *testing.T) {
			id, err := stringToID(tc.s)
			if err != nil {
				if tc.hasErr {
					return
				}
				t.Fatal(err)
			}
			assert.Equal(t, tc.id, id)
		})
	}
}
