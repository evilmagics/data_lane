package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pdf_generator/deprecated/models"
	"time"

	_ "github.com/mattn/go-adodb"

	"github.com/rs/zerolog/log"
)

// OpenConnection opens a connection to an Access database using an ODBC connection string (DSN)
// Example DSN: `Provider=Microsoft.ACE.OLEDB.12.0;Data Source=C:\path\to\db.accdb;`
// Note: Access is file-based. Keep MaxOpenConns small (1) to reduce file locking issues.
func OpenConnection(ctx context.Context, dsn string, maxOpenConns int, maxIdleConns int, connMaxLifetime time.Duration) (*Repository, error) {
	// If dsn contains just the file path, wrap it in ADODB provider string
	if !containsDriver(dsn) {
		dsn = fmt.Sprintf("Provider=Microsoft.ACE.OLEDB.12.0;Data Source=%s;", dsn)
	}

	db, err := sql.Open("adodb", dsn)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open odbc connection")
		return nil, fmt.Errorf("failed to open odbc connection: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	log.Info().Str("dsn", dsn).Msg("Opening database connection")

	err = db.PingContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Ping failed")
		_ = db.Close() // Close if ping fails
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	log.Info().Msg("Database connection established")
	return &Repository{db: db}, nil
}

func containsDriver(dsn string) bool {
	// Simple check if DSN specifies a driver
	return len(dsn) > 7 && (dsn[:7] == "Driver=" || dsn[:7] == "DRIVER=")
}

// Repository holds DB connection and config
type Repository struct {
	db *sql.DB
}

// FindTransactions performs keyset pagination using ID (recommended for large tables)
// lastID: last seen ID (0 to start from beginning)
// limit: how many rows to fetch (recommended: 500-5000 depending on memory)
// Note: Access does not support OFFSET; we use TOP N and WHERE ID > lastID
func (r *Repository) FindTransactions(ctx context.Context, limit int) ([]models.Transaction, error) {
	return r.FindTransactionsWithFilter(ctx, limit, models.TransactionFilter{})
}

func (r *Repository) FindTransactionsWithFilter(ctx context.Context, limit int, filter models.TransactionFilter) ([]models.Transaction, error) {
	log.Info().Msg("Fetching transactions with filter")

	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}
	if limit <= 0 {
		limit = 10
	}

	whereClause := "1=1"
	var args []interface{}

	if filter.StartDate != nil {
		whereClause += " AND [WAKTU] >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		whereClause += " AND [WAKTU] <= ?"
		args = append(args, *filter.EndDate)
	}
	// Time overlap logic (e.g. only transactions after certain time of day)?
	// Complex in SQL Access. For now, we might filter in memory or rely on full datetime range.

	query := fmt.Sprintf(`
		SELECT TOP %d 
			[ID], [CB], [GB], [GD], [SHIFT], [PERIODA], [IDPUL], [IDPAS], [METODA], 
			[AVC], [WAKTU], [GOL], [SERI], [STATUS], [IMAGE1], [IMAGE2]
		FROM CAPTURE
		WHERE %s
		ORDER BY [ID]
	`, limit, whereClause)

	log.Info().Str("query", query).Msg("Querying find transactions to database")

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	out := make([]models.Transaction, 0, limit)

	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(
			&t.Id,
			&t.Branch,
			&t.Gate,
			&t.Station,
			&t.Shift,
			&t.Period,
			&t.CollectorId,
			&t.PasId,
			&t.Method,
			&t.Avc,
			&t.Datetime,
			&t.Class,
			&t.Serial,
			&t.Status,
			&t.FirstImage,
			&t.SecondImage,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tx: %w", err)
		}

		// In-memory filter for TimeOverlap if needed
		// if filter.TimeOverlap != "" { ... }

		out = append(out, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return out, nil
}

// FindImages returns the two image blobs for a given ID (raw bytes). Many Access OLE Object cells
// wrap the actual image data with an OLE header. Use ExtractImageFromOLE to try to pull the
// inner image (JPEG/PNG) from the blob.
func (r *Repository) FindImages(ctx context.Context, id int) (first []byte, second []byte, err error) {
	if r == nil || r.db == nil {
		return nil, nil, errors.New("repository not initialized")
	}

	query := `SELECT IMAGE1, IMAGE2 FROM AVC WHERE ID = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	var img1, img2 []byte
	scanErr := row.Scan(&img1, &img2)
	if scanErr != nil {
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("scan images: %w", scanErr)
	}

	return img1, img2, nil
}

// ExtractImageFromOLE attempts to locate common image signatures (JPEG/PNG) inside
// an OLE/COM blob and return the inner image bytes. If no signature is found, it
// returns the original bytes unchanged.
func ExtractImageFromOLE(blob []byte) []byte {
	if len(blob) == 0 {
		return blob
	}

	// Common signatures
	jpegSig := []byte{0xFF, 0xD8, 0xFF}
	pngSig := []byte{0x89, 0x50, 0x4E, 0x47}

	// search for any of the signatures and return slice from first match
	if idx := indexOf(blob, jpegSig); idx >= 0 {
		return blob[idx:]
	}
	if idx := indexOf(blob, pngSig); idx >= 0 {
		return blob[idx:]
	}

	// fallback: return original blob
	return blob
}

// small helper to find subslice index
func indexOf(data []byte, pat []byte) int {
	if len(pat) == 0 || len(data) < len(pat) {
		return -1
	}
	for i := 0; i <= len(data)-len(pat); i++ {
		if string(data[i:i+len(pat)]) == string(pat) {
			return i
		}
	}
	return -1
}

func (r *Repository) Close() error {
	if r.db == nil {
		return nil
	}
	return r.db.Close()
}
