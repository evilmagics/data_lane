package handlers

import (
	"fmt"
	"path"
	"pdf_generator/deprecated/models"

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
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/johnfercher/maroto/v2/pkg/repository"
	"github.com/rs/zerolog/log"
)

type Generator struct {
	cfg    *entity.Config
	core   core.Maroto
	output string
}

func NewGenerator(output string) *Generator {
	fontName := "nunito-sans"
	fontPath := "./assets/fonts/nunito-sans"

	fonts, err := repository.New().
		AddUTF8Font(fontName, fontstyle.Normal, path.Join(fontPath, "nunito-sans.regular.ttf")).
		AddUTF8Font(fontName, fontstyle.Italic, path.Join(fontPath, "nunito-sans.italic.ttf")).
		AddUTF8Font(fontName, fontstyle.Bold, path.Join(fontPath, "nunito-sans.bold.ttf")).
		AddUTF8Font(fontName, fontstyle.BoldItalic, path.Join(fontPath, "nunito-sans.bold-italic.ttf")).
		Load()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load fonts")
	}

	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithTopMargin(5).
		WithBottomMargin(5).
		WithLeftMargin(5).
		WithRightMargin(5).
		WithMaxGridSize(22).
		WithCompression(true).
		WithCustomFonts(fonts).
		WithDefaultFont(&props.Font{Family: fontName}).
		WithSequentialLowMemoryMode(5).
		Build()

	return &Generator{
		cfg:    cfg,
		core:   maroto.NewMetricsDecorator(maroto.New(cfg)),
		output: output,
	}
}

func (g *Generator) Generate(transactions []models.Transaction) error {
	// g.buildTitle("Data Transaksi")
	g.buildHeader()
	g.buildBody(transactions)

	doc, err := g.core.Generate()
	if err != nil {
		return err
	}

	if err := doc.Save(g.output); err != nil {
		log.Error().Err(err).Msg("Failed to saving output")
	}

	fmt.Printf("PDF berhasil dibuat (Maroto v2): %s \n", g.output)
	fmt.Println(doc.GetReport().Normalize().String())
	return nil
}

func (g *Generator) buildTitle(title string) {
	g.core.AddRows(
		row.New(12).Add(
			col.New(12).Add(
				text.New(title, props.Text{Size: 16, Style: fontstyle.Bold, Align: align.Middle}),
			),
		),
		row.New(6).Add(col.New(12).Add(text.New("", props.Text{}))),
	)
}

func (g *Generator) buildHeader() {
	headerTextStyle := props.Text{Size: 11, Style: fontstyle.Bold, Align: align.Left, Top: 1}
	g.core.RegisterHeader(
		row.New().Add(
			col.New(16).Add(text.New("PT JASA MARGA TBK", headerTextStyle)),
			col.New(6).Add(text.New("Amplas, 28/06/2022 16:31:52",
				props.Text{Size: 10, Style: fontstyle.Normal, Align: align.Left, Top: 1, Bottom: 2},
			)),
		),
		row.New().Add(
			col.New(16).Add(text.New("CABANG : BELAWAN-MEDAN-TJ MORAWA (BALMERA)", headerTextStyle)),
			col.New(6).Add(text.New("Kabang Tol/Analis",
				props.Text{Size: 10, Style: fontstyle.Normal, Align: align.Left, Top: 0, Bottom: 4},
			)),
		),
		row.New().Add(
			col.New(16).Add(text.New("GERBANG : AMPLAS", headerTextStyle)),
			col.New(6).Add(text.New("AWM [XXX01]",
				props.Text{Size: 10, Style: fontstyle.Normal, Align: align.Left, Top: 2, Bottom: 3},
			)),
		),
	)
}

func (g *Generator) buildFooter() {}

func (g *Generator) buildBody(transactions []models.Transaction) {
	g.buildTransactionTable(transactions)
}

func (g *Generator) buildTransactionTable(transactions []models.Transaction) {
	// g.buildTransactionTableHeaders()
	g.buildTransactionTableBody(transactions)
}

func (g *Generator) buildTransactionTableHeaders() {
	textStyle := props.Text{
		Style: fontstyle.Bold,
		Size:  12,
		Align: align.Center,
	}
	borderStyle := &props.Cell{
		BorderType:      border.Full,
		BorderColor:     &props.Color{Red: 33, Green: 37, Blue: 41},
		BorderThickness: 0.2,
	}
	g.core.AddRows(
		row.New(10).Add(
			col.New(4).WithStyle(borderStyle).Add(text.New("Transaction Detail", textStyle)),
			col.New(8).WithStyle(borderStyle).Add(text.New("Captures", textStyle)),
		),
	)
}

func (g *Generator) buildTransactionTableBody(transactions []models.Transaction) {
	textSpace := 4.5
	headerTextStyle := props.Text{Size: 8, Style: fontstyle.Bold, Top: 1}
	valueTextStyle := props.Text{Size: 8, Style: fontstyle.Normal, Top: 1}
	imageStyle := props.Rect{Center: true, Percent: 95, Top: 0}
	borderStyle := &props.Cell{
		BorderType:      border.Top,
		BorderColor:     &props.Color{Red: 33, Green: 37, Blue: 41},
		BorderThickness: 0.42,
	}

	for _, t := range transactions {
		firstImage := col.New(8)
		if t.FirstImage != nil {
			firstImage.Add(image.NewFromBytes(t.FirstImage, extension.Jpg, imageStyle))
		} else {
			firstImage.Add(text.New("[No Capture]", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center}))
		}
		secondImage := col.New(8)
		if t.SecondImage != nil {
			secondImage.Add(image.NewFromBytes(t.SecondImage, extension.Jpg, imageStyle))
		} else {
			secondImage.Add(text.New("[No Capture]", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center, Top: 25}))
		}
		lastSpace := 1.0
		g.core.AddRows(
			row.New().WithStyle(borderStyle).Add(
				col.New(2).Add(
					text.New("GARDU", nextTextPropTop(headerTextStyle, textSpace, &lastSpace, true)),
					text.New("SHF/PRD", nextTextPropTop(headerTextStyle, textSpace, &lastSpace)),
					text.New("NIK PUL", nextTextPropTop(headerTextStyle, textSpace, &lastSpace)),
					text.New("NIK PAS", nextTextPropTop(headerTextStyle, textSpace, &lastSpace)),
					text.New("WAKTU", nextTextPropTop(headerTextStyle, textSpace, &lastSpace)),
					text.New("GOL/AVC", nextTextPropTop(headerTextStyle, textSpace, &lastSpace)),
					text.New("METODA", nextTextPropTop(headerTextStyle, textSpace, &lastSpace)),
					text.New("SERI", nextTextPropTop(headerTextStyle, textSpace, &lastSpace)),
					text.New("STATUS", nextTextPropTop(headerTextStyle, textSpace, &lastSpace)),
					text.New("ASAL", nextTextPropTop(headerTextStyle, textSpace, &lastSpace)),
					text.New("KARTU", nextTextPropTop(headerTextStyle, textSpace, &lastSpace)),
				),
				col.New(4).Add(
					text.New(fmt.Sprintf(": %s", t.GetStation()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace, true)),
					text.New(fmt.Sprintf(": %s / %s", t.GetShift(), t.GetPeriod()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", t.GetCollectorId()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
					text.New(fmt.Sprintf(": %s", t.GetPasId()), nextTextPropTop(valueTextStyle, textSpace, &lastSpace)),
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
