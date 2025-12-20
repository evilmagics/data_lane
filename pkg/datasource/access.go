package datasource

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-adodb"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/core/domain"
)

// Transaction represents a toll transaction
type Transaction struct {
	ID       int
	Branch   string
	Gate     string
	Station  string
	Shift    string
	Datetime string
	Class    string
	Serial   string
	Status   string
	Method   string
}

// LoadTransactions loads transactions from MS Access database
func LoadTransactions(ctx context.Context, dbPath string, filter domain.TaskFilter) ([]Transaction, error) {
	// Provider=Microsoft.ACE.OLEDB.12.0;Data Source=C:\myFolder\myAccessFile.accdb;
	dsn := fmt.Sprintf("Provider=Microsoft.ACE.OLEDB.12.0;Data Source=%s;", dbPath)

	db, err := sql.Open("adodb", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Set connection timeout
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Build query with filters
	query := `SELECT TOP 1000 [ID], [CB], [GB], [GD], [SHIFT], [WAKTU], [GOL], [SERI], [STATUS], [METODA] FROM CAPTURE WHERE 1=1`
	var args []interface{}

	if filter.Date != "" {
		query += " AND FORMAT([WAKTU], 'yyyy-MM-dd') = ?"
		args = append(args, filter.Date)
	} else if filter.RangeStart != "" && filter.RangeEnd != "" {
		query += " AND [WAKTU] BETWEEN ? AND ?"
		args = append(args, filter.RangeStart, filter.RangeEnd)
	}

	query += " ORDER BY [ID]"

	log.Debug().Str("query", query).Msg("Executing query")

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		var datetime sql.NullTime

		if err := rows.Scan(
			&t.ID,
			&t.Branch,
			&t.Gate,
			&t.Station,
			&t.Shift,
			&datetime,
			&t.Class,
			&t.Serial,
			&t.Status,
			&t.Method,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan row")
			continue
		}

		if datetime.Valid {
			t.Datetime = datetime.Time.Format("2006-01-02 15:04:05")
		}

		transactions = append(transactions, t)
	}

	return transactions, nil
}

// GetDataSourcePath constructs the path to the Access database file
// Format: {root}/{month_2_digit}{year_2_digit}/{station_id_2_digit}/{day_2_digit}{month_2_digit}{year_4_digit}.mdb
func GetDataSourcePath(rootFolder string, transactionTime time.Time, stationID string) string {
	// Format time components
	monthShort := transactionTime.Format("01")
	yearShort := transactionTime.Format("06")
	day := transactionTime.Format("02")
	yearFull := transactionTime.Format("2006")
	
	// Create folder name: MMYY (e.g., 1224)
	folderName := monthShort + yearShort
	
	// Create filename: DDMMYYYY.mdb (e.g., 20122024.mdb)
	fileName := fmt.Sprintf("%s%s%s.mdb", day, monthShort, yearFull)

	// Ensure station ID is 2 digits (e.g., "1" -> "01", "01" -> "01")
	// If stationID is not numeric, use as is (though user specified 2 digit ID)
	cleanStationID := stationID
	if len(cleanStationID) == 1 {
		cleanStationID = "0" + cleanStationID
	}
	
	return fmt.Sprintf("%s/%s/%s/%s", rootFolder, folderName, cleanStationID, fileName)
}
