package datasource

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/alexbrainman/odbc"
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
	dsn := fmt.Sprintf("Driver={Microsoft Access Driver (*.mdb, *.accdb)};DBQ=%s;", dbPath)

	db, err := sql.Open("odbc", dsn)
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
