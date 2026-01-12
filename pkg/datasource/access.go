package datasource

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/alexbrainman/odbc"
	_ "github.com/mattn/go-adodb"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/core/domain"
)

// Transaction represents a toll transaction
type Transaction struct {
	ID          int
	Branch      string
	Gate        string
	Station     string
	Shift       string
	Period      string
	CollectorID string
	PasID       string
	Avc         string
	Datetime    string
	Class       string
	Method      string
	Serial      string
	Status      string
	OriginGate  string
	CardNumber  string
	FirstImage  []byte
	SecondImage []byte
}

// LoadTransactions loads transactions from MS Access database
func LoadTransactions(ctx context.Context, dbPath string, filter domain.TaskFilter) ([]Transaction, error) {
	// Try multiple drivers as fallback
	var db *sql.DB
	var err error
	var lastErr error

	// Define driver attempts
	type driverAttempt struct {
		driver string
		dsn    string
	}

	attempts := []driverAttempt{
		{
			driver: "odbc",
			dsn:    fmt.Sprintf("Driver={Microsoft Access Driver (*.mdb, *.accdb)};Dbq=%s;", dbPath),
		},
		{
			driver: "adodb",
			dsn:    fmt.Sprintf("Provider=Microsoft.ACE.OLEDB.12.0;Data Source=%s;", dbPath),
		},
		{
			driver: "adodb",
			dsn:    fmt.Sprintf("Provider=Microsoft.Jet.OLEDB.4.0;Data Source=%s;", dbPath),
		},
	}

	for _, attempt := range attempts {
		log.Debug().Str("driver", attempt.driver).Str("dsn", attempt.dsn).Msg("Attempting to connect to MS Access")
		
		db, err = sql.Open(attempt.driver, attempt.dsn)
		if err != nil {
			log.Warn().Err(err).Str("driver", attempt.driver).Msg("Failed to open database with driver, trying next fallback")
			lastErr = err
			continue
		}

		// Set connection timeout and limits
		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetMaxOpenConns(1)

		// Create a shorter timeout for pinging
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = db.PingContext(pingCtx)
		cancel()

		if err != nil {
			db.Close()
			log.Warn().Err(err).Str("driver", attempt.driver).Msg("Failed to ping database with driver, trying next fallback")
			lastErr = err
			continue
		}

		// Success!
		log.Info().Str("driver", attempt.driver).Msg("Successfully connected to MS Access database")
		lastErr = nil
		break
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to connect to MS Access after all attempts: %w", lastErr)
	}
	defer db.Close()

	// Build query with filters
	// Use TOP clause only if limit is specified
	var query string
	if filter.Limit > 0 {
		query = fmt.Sprintf(`SELECT TOP %d [ID], [CB], [GB], [GD], [SHIFT], [PERIODA], [IDPUL], [IDPAS], [WAKTU], [GOL], [AVC], [METODA], [SERI], [STATUS], [AG], [NOKARTU], [IMAGE1], [IMAGE2] FROM CAPTURE WHERE 1=1`, filter.Limit)
	} else {
		query = `SELECT [ID], [CB], [GB], [GD], [SHIFT], [PERIODA], [IDPUL], [IDPAS], [WAKTU], [GOL], [AVC], [METODA], [SERI], [STATUS], [AG], [NOKARTU], [IMAGE1], [IMAGE2] FROM CAPTURE WHERE 1=1`
	}
	var args []interface{}

	if filter.Date != "" {
		// Daily mode: use DayStartTime to calculate transaction window
		// If DayStartTime is set (e.g., "02:00"), transactions from date+02:00 to nextDay+01:59:59
		dayStartTime := filter.DayStartTime
		if dayStartTime == "" {
			dayStartTime = "00:00"
		}
		
		// Parse the date and day start time
		startDateTime := filter.Date + " " + dayStartTime + ":00"
		
		// Calculate end time (next day at startTime - 1 second)
		parsedDate, err := time.Parse("2006-01-02", filter.Date)
		if err == nil {
			nextDay := parsedDate.AddDate(0, 0, 1)
			endDateTime := nextDay.Format("2006-01-02") + " " + dayStartTime + ":00"
			
			// Parse start and end times, then subtract 1 second from end
			startTime, _ := time.Parse("2006-01-02 15:04:05", startDateTime)
			endTime, _ := time.Parse("2006-01-02 15:04:05", endDateTime)
			endTime = endTime.Add(-time.Second)
			
			query += " AND [WAKTU] >= ? AND [WAKTU] <= ?"
			args = append(args, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))
			
			log.Debug().
				Str("date", filter.Date).
				Str("day_start_time", dayStartTime).
				Str("window_start", startTime.Format("2006-01-02 15:04:05")).
				Str("window_end", endTime.Format("2006-01-02 15:04:05")).
				Msg("Daily transaction window calculated")
		} else {
			// Fallback to simple date match if parsing fails
			query += " AND FORMAT([WAKTU], 'yyyy-MM-dd') = ?"
			args = append(args, filter.Date)
		}
	} else if filter.RangeStart != "" && filter.RangeEnd != "" {
		query += " AND [WAKTU] BETWEEN ? AND ?"
		args = append(args, filter.RangeStart, filter.RangeEnd)
	}

	query += " ORDER BY [ID]"

	log.Debug().Str("query", query).Msg("Executing query")

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	} else if rows == nil {
		return nil, fmt.Errorf("query returned no rows")
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	for rows.Next() {
		var t Transaction
		var datetime sql.NullTime

		if err := rows.Scan(
			&t.ID,
			&t.Branch,
			&t.Gate,
			&t.Station,
			&t.Shift,
			&t.Period,
			&t.CollectorID,
			&t.PasID,
			&datetime,
			&t.Class,
			&t.Avc,
			&t.Method,
			&t.Serial,
			&t.Status,
			&t.OriginGate,
			&t.CardNumber,
			&t.FirstImage,
			&t.SecondImage,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan row")
			continue
		}

		if datetime.Valid {
			t.Datetime = datetime.Time.Format("2006-01-02 15:04:05")
		}

		transactions = append(transactions, t)
	}

	log.Debug().Int("count", len(transactions)).Msg("Loaded transactions")
	return transactions, nil
}

// GetStation returns the station name or "--" if empty
func (t Transaction) GetStation() string {
	if t.Station == "" {
		return "--"
	}
	return t.Station
}

// GetShift returns the shift
func (t Transaction) GetShift() string {
	return t.Shift
}

// GetPeriod returns the period
func (t Transaction) GetPeriod() string {
	return t.Period
}

// GetCollectorID returns the collector ID or "--" if empty
func (t Transaction) GetCollectorID() string {
	if t.CollectorID == "" {
		return "--"
	}
	return t.CollectorID
}

// GetPasID returns the Pas ID or "--" if empty
func (t Transaction) GetPasID() string {
	if t.PasID == "" {
		return "--"
	}
	return t.PasID
}

// GetDatetime returns the formatted datetime
func (t Transaction) GetDatetime() string {
	// Datetime is already a string in this struct, assuming formatted in Scan
	return t.Datetime
}

// GetClass returns the class or "-" if empty
func (t Transaction) GetClass() string {
	if t.Class == "" {
		return "-"
	}
	return t.Class
}

// GetAvc returns the AVC
func (t Transaction) GetAvc() string {
	return t.Avc
}

// GetMethod returns the translated method
func (t Transaction) GetMethod() string {
	return TranslateTransactionMethod(t.Method)
}

// GetSerial returns the serial or "000000" if empty
func (t Transaction) GetSerial() string {
	if t.Serial == "" {
		return "000000"
	}
	return t.Serial
}

// GetStatus returns the status
func (t Transaction) GetStatus() string {
	return t.Status
}

// GetOriginGate returns the origin gate or "00" if empty
func (t Transaction) GetOriginGate() string {
	if t.OriginGate == "" {
		return "00"
	}
	return t.OriginGate
}

// GetCardNumber returns the card number or "--" if empty
func (t Transaction) GetCardNumber() string {
	if t.CardNumber == "" {
		return "--"
	}
	return t.CardNumber
}

var (
	transactionMethodTranslation = map[string]string{
		"PPC":  "eToll Mdr",
		"PPC0": "eToll Mdr",
		"PPC1": "eToll BRI",
		"PPC2": "eToll BNI",
		"PPC3": "eToll BTN",
		"PPC5": "eToll BCA",
		"PPC7": "eToll DKI",
	}
)

// TranslateTransactionMethod translates the transaction method code to a readable name
func TranslateTransactionMethod(method string) string {
	if m, ok := transactionMethodTranslation[method]; ok {
		return m
	}
	return method
}

// GetDataSourcePath constructs the path to the Access database file
// Format: {root}/{month_2_digit}{year_2_digit}/{gate_id_2_digit}/{day_2_digit}{month_2_digit}{year_4_digit}.mdb
func GetDataSourcePath(rootFolder string, transactionTime time.Time, gateID string) string {
	// Format time components
	monthShort := transactionTime.Format("01")
	day := transactionTime.Format("02")
	yearFull := transactionTime.Format("2006")

	// Create folder name: MM-YYYY (e.g., 12-2024)
	folderName := fmt.Sprintf("%s-%s", monthShort, yearFull)

	// Create filename: DDMMYYYY.mdb (e.g., 20122024.mdb)
	fileName := fmt.Sprintf("%s%s%s.mdb", day, monthShort, yearFull)

	// Ensure gate ID is 2 digits (e.g., "1" -> "01", "01" -> "01")
	// If gateID is not numeric, use as is (though user specified 2 digit ID)
	cleanGateID := gateID
	if len(cleanGateID) == 1 {
		cleanGateID = "0" + cleanGateID
	}

	return fmt.Sprintf("%s/%s/%s/%s", rootFolder, folderName, cleanGateID, fileName)
}
