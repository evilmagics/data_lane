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

// ProgressCallback is called to report generation progress
type ProgressCallback func(stage string, current, total int)

// GeneratePDFWithProgress creates a PDF from the given metadata with progress reporting
func GeneratePDFWithProgress(ctx context.Context, metadata domain.TaskMetadata, settingsRepo ports.SettingsRepository, onProgress ProgressCallback) (string, int64, error) {
	log.Info().Int("branch_id", metadata.BranchID).Msg("Starting PDF generation")

	// Report initial progress
	if onProgress != nil {
		onProgress("initializing", 0, 0)
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
		onProgress("loading_data", 0, 0)
	}

	dbPath := datasource.GetDataSourcePath(metadata.RootFolder, targetDate, strconv.Itoa(metadata.StationID))
	log.Debug().Str("db_path", dbPath).Msg("Using database path")
	dbPath = filepath.FromSlash(dbPath) // Ensure correct separators for Windows
	transactions, err := datasource.LoadTransactions(ctx, dbPath, metadata.Filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load transactions")
		return "", 0, err
	}

	totalTransactions := len(transactions)
	log.Info().Int("count", totalTransactions).Msg("Loaded transactions")

	if onProgress != nil {
		onProgress("loading_fonts", 0, totalTransactions)
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
		onProgress("generating", 0, totalTransactions)
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

	// Determine station name for header if available, otherwise just use ID
	stationLabel := fmt.Sprintf("GERBANG : %d", metadata.StationID)
	// Try to find a station name from the first transaction if available
	if len(transactions) > 0 {
		st := transactions[0].GetStation()
		if st != "" && st != "--" {
			stationLabel = fmt.Sprintf("GERBANG : %s", st)
		}
	}

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
			col.New(16).Add(text.New(stationLabel, headerTextStyle)),
			col.New(6).Add(text.New("AWM", // Removed [XXX01] to be generic
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
			onProgress("generating", i+1, totalTransactions)
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
					text.New(fmt.Sprintf(": %s", t.GetOriginGate()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", t.GetCardNumber()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
				),
				firstImage,
				secondImage,
			),
		)
	}

	if onProgress != nil {
		onProgress("saving", totalTransactions, totalTransactions)
	}

	// Generate output filename
	filename := formatFilename(filenameFormat, metadata)
	
	// Get absolute path for output directory
	outputDir, err := filepath.Abs("output")
	if err != nil {
		return "", 0, fmt.Errorf("failed to get absolute output path: %w", err)
	}
	
	outputPath := filepath.Join(outputDir, filename+".pdf")

	// Ensure output directory exists
	os.MkdirAll(outputDir, 0755)

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

	if onProgress != nil {
		onProgress("completed", totalTransactions, totalTransactions)
	}

	log.Info().Str("output", outputPath).Int64("size", info.Size()).Msg("PDF generated")
	return outputPath, info.Size(), nil
}

// GeneratePDF creates a PDF from the given metadata (legacy wrapper without progress)
func GeneratePDF(ctx context.Context, metadata domain.TaskMetadata, settingsRepo ports.SettingsRepository) (string, int64, error) {
	return GeneratePDFWithProgress(ctx, metadata, settingsRepo, nil)
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
