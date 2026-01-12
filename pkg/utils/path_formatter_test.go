package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatPath(t *testing.T) {
	params := PathParams{
		Time:      time.Date(2024, 12, 25, 14, 30, 05, 0, time.UTC),
		BranchID:  1,
		StationID: 5,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Full path with dates and IDs",
			template: "{YYYY}/{MM}/{DD}/{BranchID}/{StationID}/data.mdb",
			expected: "2024/12/25/01/05/data.mdb",
		},
		{
			name:     "Short year and alias GateID",
			template: "{YY}{MM}/{GateID}/file.pdf",
			expected: "2412/05/file.pdf",
		},
		{
			name:     "Date and Time keys",
			template: "report_{Date}_{Time}.pdf",
			expected: "report_20241225_143005.pdf",
		},
		{
			name:     "Legacy support",
			template: "{branch_id}_{gate_id}_{date}",
			expected: "01_05_20241225",
		},
		{
			name:     "No placeholders",
			template: "static/path/file.txt",
			expected: "static/path/file.txt",
		},
		{
			name:     "Padding for 2-digit IDs",
			template: "{BranchID}_{StationID}",
			expected: "01_05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FormatPath(tt.template, params)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
