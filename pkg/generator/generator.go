package generator

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	maroto "github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/johnfercher/maroto/v2/pkg/repository"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
	"pdf_generator/pkg/datasource"
)

// DefaultOutputDir is the default directory for PDF output files
const DefaultOutputDir = "output"

// ProgressCallback is called to report generation progress
type ProgressCallback func(stage string, current, total int)

// GeneratePDFWithProgress creates a PDF from the given metadata with progress reporting
func GeneratePDFWithProgress(ctx context.Context, metadata domain.TaskMetadata, settingsRepo ports.SettingsRepository, gateRepo ports.GateRepository, onProgress ProgressCallback) (string, int64, error) {
	log.Info().Int("branch_id", metadata.BranchID).Msg("Starting PDF generation")

	// Report initial progress
	if onProgress != nil {
		onProgress("Initializing settings", 0, 0)
	}

	// Load settings
	branchName := getSettingOrDefault(ctx, settingsRepo, domain.SettingBranchName, strconv.Itoa(metadata.BranchID))
	if name, ok := metadata.Settings["branch_name"]; ok && name != "" {
		branchName = name
	}

	company := getSettingOrDefault(ctx, settingsRepo, domain.SettingManagementCompany, "PT JASA MARGA TBK")
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

	// Get day start time for daily transaction window
	dayStartTime := getSettingOrDefault(ctx, settingsRepo, domain.SettingTimeOverlap, "00:00")
	if dst, ok := metadata.Settings["day_start_time"]; ok && dst != "" {
		dayStartTime = dst
	}

	// Get analyzer operator name for PDF header
	analyzerOperatorName := getSettingOrDefault(ctx, settingsRepo, domain.SettingAnalyzerOperatorName, "Analyzer Operator")
	if aon, ok := metadata.Settings["analyzer_operator_name"]; ok && aon != "" {
		analyzerOperatorName = aon
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

	if onProgress != nil {
		onProgress("Connecting to database", 0, 0)
	}

	dbPath := datasource.GetDataSourcePath(metadata.RootFolder, targetDate, strconv.Itoa(metadata.GateID))
	dbPath = filepath.FromSlash(dbPath) // Ensure correct separators for Windows

	// Check datasource from filepath while running the task
	if _, err := os.Stat(dbPath); err != nil {
		log.Info().Str("path", dbPath).Msg("Datasource file not found while running task")
	} else {
		log.Info().Str("path", dbPath).Msg("Datasource file found while running task")
	}

	// Populate DayStartTime in filter for daily queries
	filter := metadata.Filter
	if filter.Date != "" && filter.DayStartTime == "" {
		filter.DayStartTime = dayStartTime
	}

	transactions, err := datasource.LoadTransactions(ctx, dbPath, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load transactions")
		return "", 0, err
	}

	totalTransactions := len(transactions)
	log.Info().Int("count", totalTransactions).Msg("Loaded transactions")

	// Build gate name lookup map
	gateNameMap := make(map[int]string)
	if gateRepo != nil {
		gates, err := gateRepo.List(ctx)
		if err == nil {
			for _, g := range gates {
				gateNameMap[g.ID] = g.Name
			}
		}
	}

	// Helper function to get gate name by ID
	getGateName := func(gateID int) string {
		if name, ok := gateNameMap[gateID]; ok {
			return name
		}
		return strconv.Itoa(gateID) // Fallback to ID if name not found
	}

	if onProgress != nil {
		onProgress("Loading fonts", 0, totalTransactions)
	}

	// Load Fonts
	fontName := "nunito-sans"
	fontPath := "./assets/fonts/nunito-sans"

	// Validate font path existence
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		log.Warn().Str("path", fontPath).Msg("Font directory not found, using default fonts")
	}

	var fonts []*entity.CustomFont
	var loadErr error

	fonts, loadErr = repository.New().
		AddUTF8Font(fontName, fontstyle.Normal, path.Join(fontPath, "nunito-sans.regular.ttf")).
		AddUTF8Font(fontName, fontstyle.Italic, path.Join(fontPath, "nunito-sans.italic.ttf")).
		AddUTF8Font(fontName, fontstyle.Bold, path.Join(fontPath, "nunito-sans.bold.ttf")).
		AddUTF8Font(fontName, fontstyle.BoldItalic, path.Join(fontPath, "nunito-sans.bold-italic.ttf")).
		Load()

	if loadErr != nil {
		log.Error().Err(loadErr).Msg("Failed to load custom fonts, falling back to default")
		fonts = nil // Ensure nil if error
	}

	if onProgress != nil {
		onProgress("Building PDF header", 0, totalTransactions)
	}

	// Create PDF Config
	builder := config.NewBuilder().
		WithPageSize(getPageSize(pageSize)).
		WithTopMargin(5).
		WithBottomMargin(5).
		WithLeftMargin(5).
		WithRightMargin(5).
		WithMaxGridSize(22).
		WithCompression(true).
		WithSequentialLowMemoryMode(5)

	if fonts != nil {
		builder.WithCustomFonts(fonts).
			WithDefaultFont(&props.Font{Family: fontName})
	}

	cfg := builder.Build()

	m := maroto.New(cfg)

	// Header Logic adapted from deprecated
	headerTextStyle := props.Text{Size: 11, Style: fontstyle.Bold, Align: align.Left, Top: 1}

	// Use gate name from lookup for header
	gateLabel := fmt.Sprintf("GERBANG : %s", getGateName(metadata.GateID))

	dateStr := time.Now().Format("02/01/2006 15:04:05")

	m.RegisterHeader(
		row.New().Add(
			col.New(16).Add(text.New(company, headerTextStyle)),
			col.New(6).Add(text.New(dateStr,
				props.Text{Size: 10, Style: fontstyle.Normal, Align: align.Left, Top: 1, Bottom: 2},
			)),
		),
		row.New().Add(
			col.New(16).Add(text.New(fmt.Sprintf("CABANG : %s", branchName), headerTextStyle)),
			col.New(6).Add(text.New("Kabang Tol/Analis",
				props.Text{Size: 10, Style: fontstyle.Normal, Align: align.Left, Top: 0, Bottom: 4},
			)),
		),
		row.New().Add(
			col.New(16).Add(text.New(gateLabel, headerTextStyle)),
			col.New(6).Add(text.New(analyzerOperatorName,
				props.Text{Size: 10, Style: fontstyle.Normal, Align: align.Left, Top: 2, Bottom: 3},
			)),
		),
	)

	// Body Logic (Transactions)
	textSpace := 4.5
	bodyCheckTextStyle := props.Text{Size: 8, Style: fontstyle.Bold, Top: 1}
	valueTextStyle := props.Text{Size: 8, Style: fontstyle.Normal, Top: 1}
	imageStyle := props.Rect{Center: true, Percent: 95, Top: 0}
	borderStyle := &props.Cell{
		BorderType:      border.Top,
		BorderColor:     &props.Color{Red: 33, Green: 37, Blue: 41},
		BorderThickness: 0.42,
	}

	for i, t := range transactions {
		// Report progress for each transaction
		if onProgress != nil {
			onProgress(fmt.Sprintf("Appending transaction %d of %d", i+1, totalTransactions), i+1, totalTransactions)
		}

		firstImage := col.New(8)
		if len(t.FirstImage) > 0 {
			firstImage.Add(image.NewFromBytes(t.FirstImage, extension.Jpg, imageStyle))
		} else {
			firstImage.Add(text.New("[No Capture]", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center}))
		}

		secondImage := col.New(8)
		if len(t.SecondImage) > 0 {
			secondImage.Add(image.NewFromBytes(t.SecondImage, extension.Jpg, imageStyle))
		} else {
			// Align logic from deprecated: Top: 25 to push it down?
			secondImage.Add(text.New("[No Capture]", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center, Top: 25}))
		}

		lastSpace := 1.0

		m.AddRows(
			row.New().WithStyle(borderStyle).Add(
				col.New(2).Add(
					text.New("GARDU", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace, true)),
					text.New("SHF/PRD", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace)),
					text.New("NIK PUL", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace)),
					text.New("NIK PAS", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace)),
					text.New("WAKTU", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace)),
					text.New("GOL/AVC", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace)),
					text.New("METODA", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace)),
					text.New("SERI", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace)),
					text.New("STATUS", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace)),
					text.New("ASAL", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace)),
					text.New("KARTU", nextTextPropTop(bodyCheckTextStyle, textSpace, &lastSpace)),
				),
				col.New(4).Add(
					text.New(fmt.Sprintf(": %s", t.GetStation()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace, true)),
					text.New(fmt.Sprintf(": %s / %s", t.GetShift(), t.GetPeriod()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", t.GetCollectorID()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", t.GetPasID()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", t.GetDatetime()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s / %s", t.GetClass(), t.GetAvc()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", t.GetMethod()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", t.GetSerial()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", t.GetStatus()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", func() string {
						// Look up origin gate name by ID
						originGateStr := t.GetOriginGate()
						if originGateID, err := strconv.Atoi(originGateStr); err == nil {
							return getGateName(originGateID)
						}
						return originGateStr // Return original if not a valid number
					}()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", t.GetCardNumber()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
				),
				firstImage,
				secondImage,
			),
		)
	}

	// Report completion of transaction appending
	if onProgress != nil {
		onProgress("All transactions appended", totalTransactions, totalTransactions)
	}

	if onProgress != nil {
		onProgress("Rendering PDF document", totalTransactions, totalTransactions)
	}

	// Generate output filename
	filename := formatFilename(filenameFormat, metadata)

	// Get absolute path for output directory
	outputDir, err := filepath.Abs(DefaultOutputDir)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get absolute output path: %w", err)
	}

	outputPath := filepath.Join(outputDir, filename+".pdf")

	// Generate document
	doc, err := m.Generate()
	if err != nil {
		return "", 0, err
	}

	if onProgress != nil {
		onProgress("Writing file to disk", totalTransactions, totalTransactions)
	}

	if err := doc.Save(outputPath); err != nil {
		return "", 0, err
	}

	// Get file size
	info, err := os.Stat(outputPath)
	if err != nil {
		return outputPath, 0, nil
	}

	if onProgress != nil {
		onProgress("Completed", totalTransactions, totalTransactions)
	}

	log.Info().Str("output", outputPath).Int64("size", info.Size()).Msg("PDF generated")
	return outputPath, info.Size(), nil
}

// GeneratePDF creates a PDF from the given metadata (legacy wrapper without progress)
func GeneratePDF(ctx context.Context, metadata domain.TaskMetadata, settingsRepo ports.SettingsRepository, gateRepo ports.GateRepository) (string, int64, error) {
	return GeneratePDFWithProgress(ctx, metadata, settingsRepo, gateRepo, nil)
}

// GenerateMultiDatePDF handles date range generation, producing one PDF per date
// Returns comma-separated output paths and total size across all files
func GenerateMultiDatePDF(ctx context.Context, metadata domain.TaskMetadata, settingsRepo ports.SettingsRepository, gateRepo ports.GateRepository, onProgress ProgressCallback) (string, int64, error) {
	// Check if this is a date range request
	if metadata.Filter.RangeStart == "" || metadata.Filter.RangeEnd == "" {
		// Single date mode - use existing generator
		return GeneratePDFWithProgress(ctx, metadata, settingsRepo, gateRepo, onProgress)
	}

	// Parse date range
	startDate, err := time.Parse("2006-01-02", metadata.Filter.RangeStart)
	if err != nil {
		return "", 0, fmt.Errorf("invalid range_start date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", metadata.Filter.RangeEnd)
	if err != nil {
		return "", 0, fmt.Errorf("invalid range_end date: %w", err)
	}

	// Calculate number of days
	days := int(endDate.Sub(startDate).Hours()/24) + 1
	if days <= 0 {
		return "", 0, fmt.Errorf("invalid date range: end date must be after start date")
	}

	log.Info().
		Str("range_start", metadata.Filter.RangeStart).
		Str("range_end", metadata.Filter.RangeEnd).
		Int("days", days).
		Msg("Starting multi-date PDF generation")

	if onProgress != nil {
		onProgress(fmt.Sprintf("Processing date range: %d days", days), 0, days)
	}

	var outputPaths []string
	var totalSize int64

	// Iterate through each date
	for i := 0; i < days; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		dateStr := currentDate.Format("2006-01-02")

		if onProgress != nil {
			onProgress(fmt.Sprintf("Processing date %d of %d: %s", i+1, days, dateStr), i, days)
		}

		// Create a copy of metadata with single date filter
		singleDayMetadata := metadata
		singleDayMetadata.Filter = domain.TaskFilter{
			Date:              dateStr,
			DayStartTime:      metadata.Filter.DayStartTime,
			TransactionStatus: metadata.Filter.TransactionStatus,
			Limit:             metadata.Filter.Limit,
		}

		// Create per-date progress callback that prefixes with date info
		perDateProgress := func(stage string, current, total int) {
			if onProgress != nil {
				prefixedStage := fmt.Sprintf("[%s] %s", dateStr, stage)
				onProgress(prefixedStage, current, total)
			}
		}

		// Generate PDF for this single date
		output, size, err := GeneratePDFWithProgress(ctx, singleDayMetadata, settingsRepo, gateRepo, perDateProgress)
		if err != nil {
			log.Warn().Err(err).Str("date", dateStr).Msg("Failed to generate PDF for date, skipping")
			continue // Skip failed dates but continue with others
		}

		outputPaths = append(outputPaths, output)
		totalSize += size

		log.Info().
			Str("date", dateStr).
			Str("output", output).
			Int64("size", size).
			Msg("Generated PDF for date")
	}

	if len(outputPaths) == 0 {
		return "", 0, fmt.Errorf("no PDFs were generated for the date range")
	}

	if onProgress != nil {
		onProgress(fmt.Sprintf("Completed: %d PDFs generated", len(outputPaths)), days, days)
	}

	// Return comma-separated paths for multiple files
	return strings.Join(outputPaths, ","), totalSize, nil
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
	result = strings.ReplaceAll(result, "{station_id}", strconv.Itoa(metadata.GateID))

	// Use transaction filter date for {date}, not generation time
	dateStr := now.Format("20060102") // Default to current date
	if metadata.Filter.Date != "" {
		// Parse filter date and reformat
		if parsed, err := time.Parse("2006-01-02", metadata.Filter.Date); err == nil {
			dateStr = parsed.Format("20060102")
		}
	} else if metadata.Filter.RangeStart != "" {
		// For range mode, use the start date
		if parsed, err := time.Parse("2006-01-02", metadata.Filter.RangeStart[:10]); err == nil {
			dateStr = parsed.Format("20060102")
		}
	}
	result = strings.ReplaceAll(result, "{date}", dateStr)

	// Keep {time} as current time for uniqueness in filenames
	result = strings.ReplaceAll(result, "{time}", now.Format("150405"))
	return result
}

func nextTextPropTop(p props.Text, space float64, last *float64, reset ...bool) props.Text {
	if len(reset) > 0 && reset[0] {
		p.Top = 1.0
		*last = 1.0
		return p
	}
	p.Top = space + (*last)
	*last = p.Top
	return p
}
