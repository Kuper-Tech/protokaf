package calldata

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func Test_randomStringWithCharset(t *testing.T) {
	tests := []struct {
		charset string
		length  int
	}{
		{"", 10},
		{"abc", 10},
		{"abc", 0},
		{"abc", -1},
		{"😃😄😁", 0},
		{"абвabc", 6},
	}
	for _, tt := range tests {
		s := randomStringWithCharset(tt.charset, tt.length)
		c := utf8.RuneCountInString(s)

		if tt.length <= 0 {
			assert.Greater(t, c, 0)
		} else {
			assert.Equal(t, tt.length, c)
		}

		cs := tt.charset
		if cs == "" {
			cs = defaultCharset
		}

		for _, r := range s {
			assert.Contains(t, []rune(cs), r)
		}
	}
}
