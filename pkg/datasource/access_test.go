package datasource

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDataSourcePath(t *testing.T) {
	rootFolder := "d:/data"
	txTime := time.Date(2024, 2, 5, 10, 0, 0, 0, time.UTC)

	// Test case 1: Custom format
	format := "{MM}{YY}/{StationID}/{DD}{MM}{YYYY}.mdb"
	expected := "d:/data/0224/01/05022024.mdb"
	actual := GetDataSourcePath(format, rootFolder, txTime, 0, 0, 1) // branchID=0, gateID=0, stationID=1
	
	assert.Equal(t, expected, actual)

	// Test case 2: Double digit station ID and different format
	stationID2 := 10
	format2 := "{YYYY}/{MM}/{DD}/{StationID}/data.mdb"
	expected2 := "d:/data/2024/02/05/10/data.mdb"
	actual2 := GetDataSourcePath(format2, rootFolder, txTime, 0, 0, stationID2)
	assert.Equal(t, expected2, actual2)
	
	// Test case 3: Including BranchID
	branchID := 2
	format3 := "{BranchID}/{StationID}/{Date}.mdb"
	expected3 := "d:/data/02/10/20231225.mdb"
	txTime3 := time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC)
	actual3 := GetDataSourcePath(format3, rootFolder, txTime3, branchID, 0, stationID2)
	assert.Equal(t, expected3, actual3)
}
