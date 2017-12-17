package cli

import (
	"os"
	"strings"

	"github.com/tealeg/xlsx"
	"github.com/uphy/coin-trade-history/services"
)

type Writer interface {
	Write(data *services.TradeData) error
	Close() error
}

type ExcelWriter struct {
	file      string
	excelFile *xlsx.File
	sheet     *xlsx.Sheet
	total     float64
}

func NewExcelWriter(output string) Writer {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Trades")
	if err != nil {
		panic(err)
	}
	sheet.AutoFilter = &xlsx.AutoFilter{
		TopLeftCell:     "A1",
		BottomRightCell: "H1",
	}
	sheet.SheetViews = []xlsx.SheetView{
		xlsx.SheetView{
			Pane: &xlsx.Pane{
				XSplit:      1,
				YSplit:      1,
				ActivePane:  "bottomRight",
				TopLeftCell: "B2",
				State:       "frozen",
			},
		},
	}
	columns := []struct {
		Title string
		Width float64
	}{
		{"Time", 20},
		{"Service", 15},
		{"Currency", 15},
		{"Action", 15},
		{"Price", 15},
		{"Amount", 20},
		{"Fee", 20},
		{"Profit", 20},
		{"Total", 25},
		{"Remarks", 30},
	}
	header := sheet.AddRow()
	for i, col := range columns {
		cell := header.AddCell()
		style := cell.GetStyle()
		style.Font.Bold = true
		style.Alignment = xlsx.Alignment{Horizontal: "center"}
		style.ApplyAlignment = true
		style.Fill.PatternType = "solid"
		style.Fill.FgColor = "EEEEEE"
		style.ApplyFill = true
		cell.SetValue(col.Title)
		sheet.Col(i).Width = col.Width
	}
	return &ExcelWriter{output, file, sheet, 0}
}

func (e *ExcelWriter) Write(data *services.TradeData) error {
	e.total += data.Profit
	row := e.sheet.AddRow()
	row.AddCell().SetValue(data.Time.Format("2006/01/02 15:04:05.000"))
	row.AddCell().SetValue(data.ServiceName)
	row.AddCell().SetValue(data.CurrencyPair)
	row.AddCell().SetValue(data.Action.String())
	row.AddCell().SetValue(data.Price)
	row.AddCell().SetValue(data.Amount)
	row.AddCell().SetValue(data.Fee)
	row.AddCell().SetValue(data.Profit)
	row.AddCell().SetValue(e.total)
	if data.Remarks != nil && len(data.Remarks) > 0 {
		row.AddCell().SetValue(strings.Join(data.Remarks, "\n"))
	}
	return nil
}

func (e *ExcelWriter) Close() error {
	out, err := os.Create(e.file)
	if err != nil {
		return err
	}
	return e.excelFile.Write(out)
}
