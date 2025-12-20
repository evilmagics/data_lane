package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	maroto "github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
	"pdf_generator/pkg/datasource"
)

// GeneratePDF creates a PDF from the given metadata
func GeneratePDF(ctx context.Context, metadata domain.TaskMetadata, settingsRepo ports.SettingsRepository) (string, int64, error) {
	log.Info().Int("branch_id", metadata.BranchID).Msg("Starting PDF generation")

	// Load settings
	branchName := getSettingOrDefault(ctx, settingsRepo, domain.SettingBranchName, strconv.Itoa(metadata.BranchID))
	if name, ok := metadata.Settings["branch_name"]; ok && name != "" {
		branchName = name
	}

	company := getSettingOrDefault(ctx, settingsRepo, domain.SettingManagementCompany, "")
	if c, ok := metadata.Settings["management_company"]; ok && c != "" {
		company = c
	}

	pageSize := getSettingOrDefault(ctx, settingsRepo, domain.SettingPageSize, "A4")
	if ps, ok := metadata.Settings["page_size"]; ok && ps != "" {
		pageSize = ps
	}

	filenameFormat := getSettingOrDefault(ctx, settingsRepo, domain.SettingOutputFilenameFormat, "{branch_id}_{date}")
	if ff, ok := metadata.Settings["output_filename_format"]; ok && ff != "" {
		filenameFormat = ff
	}

	// Connect to Access database
	var targetDate time.Time
	var err error

	if metadata.Filter.Date != "" {
		targetDate, err = time.Parse("2006-01-02", metadata.Filter.Date)
	} else if metadata.Filter.RangeStart != "" {
		targetDate, err = time.Parse("2006-01-02", metadata.Filter.RangeStart)
	} else {
		targetDate = time.Now()
	}

	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse date for data source path, using current time")
		targetDate = time.Now()
	}

	dbPath := datasource.GetDataSourcePath(metadata.RootFolder, targetDate, strconv.Itoa(metadata.StationID))
	dbPath = filepath.FromSlash(dbPath) // Ensure correct separators for Windows
	transactions, err := datasource.LoadTransactions(ctx, dbPath, metadata.Filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load transactions")
		return "", 0, err
	}

	log.Info().Int("count", len(transactions)).Msg("Loaded transactions")

	// Create PDF
	cfg := config.NewBuilder().
		WithPageSize(getPageSize(pageSize)).
		WithTopMargin(10).
		WithBottomMargin(10).
		WithLeftMargin(10).
		WithRightMargin(10).
		Build()

	m := maroto.New(cfg)

	// Header
	m.AddRows(
		row.New(10).Add(
			col.New(12).Add(
				text.New(company, props.Text{Size: 14, Style: fontstyle.Bold, Align: align.Center}),
			),
		),
		row.New(8).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Branch: %s", branchName), props.Text{Size: 11, Align: align.Center}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Date: %s", metadata.Filter.Date), props.Text{Size: 10, Align: align.Center}),
			),
		),
	)

	// Table of transactions (simplified)
	for _, tx := range transactions {
		m.AddRows(
			row.New(8).Add(
				col.New(3).Add(text.New(tx.Datetime, props.Text{Size: 8})),
				col.New(2).Add(text.New(tx.Gate, props.Text{Size: 8})),
				col.New(2).Add(text.New(tx.Class, props.Text{Size: 8})),
				col.New(3).Add(text.New(tx.Serial, props.Text{Size: 8})),
				col.New(2).Add(text.New(tx.Status, props.Text{Size: 8})),
			),
		)
	}

	// Generate output filename
	filename := formatFilename(filenameFormat, metadata)
	outputPath := filepath.Join("output", filename+".pdf")

	// Ensure output directory exists
	os.MkdirAll("output", 0755)

	// Generate document
	doc, err := m.Generate()
	if err != nil {
		return "", 0, err
	}

	if err := doc.Save(outputPath); err != nil {
		return "", 0, err
	}

	// Get file size
	info, err := os.Stat(outputPath)
	if err != nil {
		return outputPath, 0, nil
	}

	log.Info().Str("output", outputPath).Int64("size", info.Size()).Msg("PDF generated")
	return outputPath, info.Size(), nil
}

func getSettingOrDefault(ctx context.Context, repo ports.SettingsRepository, key, defaultVal string) string {
	setting, err := repo.Get(ctx, key)
	if err != nil || setting == nil {
		return defaultVal
	}
	return setting.Value
}

func getPageSize(size string) pagesize.Type {
	switch strings.ToUpper(size) {
	case "A3":
		return pagesize.A3
	case "A5":
		return pagesize.A5
	case "LETTER":
		return pagesize.Letter
	default:
		return pagesize.A4
	}
}

func formatFilename(format string, metadata domain.TaskMetadata) string {
	now := time.Now()
	result := format
	result = strings.ReplaceAll(result, "{branch_id}", strconv.Itoa(metadata.BranchID))
	result = strings.ReplaceAll(result, "{gate_id}", strconv.Itoa(metadata.GateID))
	result = strings.ReplaceAll(result, "{station_id}", strconv.Itoa(metadata.StationID))
	result = strings.ReplaceAll(result, "{date}", now.Format("20060102"))
	result = strings.ReplaceAll(result, "{time}", now.Format("150405"))
	return result
}
