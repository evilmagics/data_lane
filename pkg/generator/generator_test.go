package generator

import (
	"testing"

	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/stretchr/testify/assert"
)

func TestGetPageSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected pagesize.Type
	}{
		{"A3", "A3", pagesize.A3},
		{"A5", "A5", pagesize.A5},
		{"Letter", "LETTER", pagesize.Letter},
		{"Default", "Unknown", pagesize.A4},
		{"Empty", "", pagesize.A4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPageSize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNextTextPropTop(t *testing.T) {
	last := 0.0
	textProp := props.Text{}

	// First call with reset
	res := nextTextPropTop(textProp, 5.0, &last, true)
	assert.Equal(t, 1.0, res.Top)
	assert.Equal(t, 1.0, last)

	// Second call
	res = nextTextPropTop(textProp, 5.0, &last)
	// Should be space (5.0) + last (1.0) = 6.0
	assert.Equal(t, 6.0, res.Top)
	assert.Equal(t, 6.0, last)
}
