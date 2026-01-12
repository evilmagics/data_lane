package utils

import (
	"strconv"
	"strings"
	"time"
)

// PathParams contains parameters for path formatting
type PathParams struct {
	Time      time.Time
	BranchID  int
	StationID int
	GateID    int // Alias for StationID
}

// FormatPath replaces placeholders in a template string with actual values
// Supported keys:
// - {YYYY}: 4-digit year
// - {YY}: 2-digit year
// - {MM}: 2-digit month
// - {DD}: 2-digit day
// - {BranchID}: 2-digit branch ID
// - {StationID}: 2-digit station ID
// - {GateID}: 2-digit gate ID (alias for StationID)
// - {Date}: Date in YYYYMMDD format
// - {Time}: Time in HHMMSS format
func FormatPath(template string, params PathParams) string {
	result := template

	// Date-time components
	result = strings.ReplaceAll(result, "{YYYY}", params.Time.Format("2006"))
	result = strings.ReplaceAll(result, "{YY}", params.Time.Format("06"))
	result = strings.ReplaceAll(result, "{MM}", params.Time.Format("01"))
	result = strings.ReplaceAll(result, "{DD}", params.Time.Format("02"))
	result = strings.ReplaceAll(result, "{Date}", params.Time.Format("20060102"))
	result = strings.ReplaceAll(result, "{Time}", params.Time.Format("150405"))

	// ID components with 2-digit padding
	padID := func(id int) string {
		s := strconv.Itoa(id)
		if len(s) == 1 {
			return "0" + s
		}
		return s
	}

	branchIDStr := padID(params.BranchID)
	stationIDStr := padID(params.StationID)
	// If GateID is explicitly provided (not 0), use it, otherwise use StationID
	gateIDStr := stationIDStr
	if params.GateID != 0 {
		gateIDStr = padID(params.GateID)
	}

	result = strings.ReplaceAll(result, "{BranchID}", branchIDStr)
	result = strings.ReplaceAll(result, "{StationID}", stationIDStr)
	result = strings.ReplaceAll(result, "{GateID}", gateIDStr)

	// Legacy support (lowercase/underscore versions if any)
	result = strings.ReplaceAll(result, "{branch_id}", branchIDStr)
	result = strings.ReplaceAll(result, "{station_id}", stationIDStr)
	result = strings.ReplaceAll(result, "{gate_id}", gateIDStr)
	result = strings.ReplaceAll(result, "{date}", params.Time.Format("20060102"))
	result = strings.ReplaceAll(result, "{DATE}", params.Time.Format("20060102"))
	result = strings.ReplaceAll(result, "{time}", params.Time.Format("150405"))

	return result
}
