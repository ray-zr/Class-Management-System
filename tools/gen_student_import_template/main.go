package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func main() {
	out := filepath.FromSlash("docs/frontend/templates/students_import_template.xlsx")
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		panic(err)
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetName(sheet, "Students")
	sheet = "Students"

	_ = f.SetCellValue(sheet, "A1", "StudentNo")
	_ = f.SetCellValue(sheet, "B1", "Name")
	_ = f.SetCellValue(sheet, "C1", "Gender")
	_ = f.SetCellValue(sheet, "D1", "Phone")
	_ = f.SetCellValue(sheet, "E1", "Position")

	_ = f.SetCellValue(sheet, "A2", "20260001")
	_ = f.SetCellValue(sheet, "B2", "张三")
	_ = f.SetCellValue(sheet, "C2", "男")
	_ = f.SetCellValue(sheet, "D2", "13800000000")
	_ = f.SetCellValue(sheet, "E2", "班长")

	_ = f.SetCellValue(sheet, "A3", "20260002")
	_ = f.SetCellValue(sheet, "B3", "李四")
	_ = f.SetCellValue(sheet, "C3", "女")
	_ = f.SetCellValue(sheet, "D3", "")
	_ = f.SetCellValue(sheet, "E3", "")

	_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	_ = f.SetColWidth(sheet, "A", "A", 14)
	_ = f.SetColWidth(sheet, "B", "B", 12)
	_ = f.SetColWidth(sheet, "C", "C", 8)
	_ = f.SetColWidth(sheet, "D", "D", 14)
	_ = f.SetColWidth(sheet, "E", "E", 10)

	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	_ = f.SetCellStyle(sheet, "A1", "E1", headerStyle)

	if err := f.SaveAs(out); err != nil {
		panic(err)
	}
	fmt.Println(out)
}
