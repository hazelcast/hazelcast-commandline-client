package plug

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestValidateName(t *testing.T) {
	testCases := []struct {
		name  string
		valid bool
	}{
		{name: "", valid: false},
		{name: ":", valid: false},
		{name: ":g", valid: false},
		{name: "map:", valid: false},
		{name: "map:get", valid: true},
		{name: "Map:get", valid: false},
		{name: "map:get2", valid: false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.valid, validName(tc.name))
		})
	}
}
