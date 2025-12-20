package datasource

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDataSourcePath(t *testing.T) {
	rootFolder := "d:/data"
	stationID := "1" // Should be padded to "01"
	
	// Test case 1: Single digit day and month (if applicable, though month is usually 2 digits)
	// Date: 2024-02-05
	txTime := time.Date(2024, 2, 5, 10, 0, 0, 0, time.UTC)
	
	expected := "d:/data/022024/01/05022024.mdb"
	actual := GetDataSourcePath(rootFolder, txTime, stationID)
	
	assert.Equal(t, expected, actual)

	// Test case 2: Double digit station ID
	stationID2 := "10"
	expected2 := "d:/data/022024/10/05022024.mdb"
	actual2 := GetDataSourcePath(rootFolder, txTime, stationID2)
	assert.Equal(t, expected2, actual2)
	
	// Test case 3: Different month/year
	// Date: 2023-12-25
	txTime3 := time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC)
	expected3 := "d:/data/122023/10/25122023.mdb"
	actual3 := GetDataSourcePath(rootFolder, txTime3, stationID2)
	assert.Equal(t, expected3, actual3)
}
