package datasource

import (
	"testing"

	"pdf_generator/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestBuildQuery(t *testing.T) {
	gateID := 5
	
	tests := []struct {
		name     string
		filter   domain.TaskFilter
		contains []string
		args     []interface{}
	}{
		{
			name: "Basic date filter",
			filter: domain.TaskFilter{
				Date: "2024-02-05",
			},
			contains: []string{"WHERE 1=1 AND [WAKTU] >= ? AND [WAKTU] <= ?"},
		},
		{
			name: "GateID filter",
			filter: domain.TaskFilter{
				GateID: &gateID,
			},
			contains: []string{"AND [GB] = ?"},
			args:     []interface{}{gateID},
		},
		{
			name: "OriginGateIDs filter",
			filter: domain.TaskFilter{
				OriginGateIDs: []int{1, 2, 3},
			},
			contains: []string{"AND [AG] IN (?,?,?)"},
			args:     []interface{}{1, 2, 3},
		},
		{
			name: "TransactionStatus filter",
			filter: domain.TaskFilter{
				TransactionStatus: "Completed",
			},
			contains: []string{"AND [STATUS] = ?"},
			args:     []interface{}{"Completed"},
		},
		{
			name: "Mixed filters",
			filter: domain.TaskFilter{
				GateID:            &gateID,
				OriginGateIDs:     []int{10},
				TransactionStatus: "Failed",
			},
			contains: []string{"AND [GB] = ?", "AND [AG] IN (?)", "AND [STATUS] = ?"},
			args:     []interface{}{gateID, 10, "Failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := buildQuery(tt.filter)
			for _, c := range tt.contains {
				assert.Contains(t, query, c)
			}
			// Skip date args check as they are dynamic in buildQuery based on DayStartTime
			if len(tt.args) > 0 {
				// Find args that match our expected ones
				for _, expectedArg := range tt.args {
					assert.Contains(t, args, expectedArg)
				}
			}
		})
	}
}
