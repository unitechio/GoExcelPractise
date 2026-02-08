package exporter

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// Style cho Tiêu đề báo cáo chính (BÁO CÁO...)
func createTitleStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Color:  "000000",
			Family: "Times New Roman",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
}

// Style cho Tên công ty/Quốc hiệu (In đậm)
func createTopHeaderBoldStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14, // Updated to 14
			Color:  "000000",
			Family: "Times New Roman",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   false, // Disable wrap text to prevent ugly stacking
		},
	})
}

// Style cho thông tin phụ (Phòng kinh doanh, Độc lập tự do..., Ngày tháng)
func createTopHeaderNormalStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   14, // Updated to 14
			Color:  "000000",
			Family: "Times New Roman",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   false, // Disable wrap text
		},
	})
}

// Style cho dòng Ngày tháng (In nghiêng)
func createTopHeaderItalicStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Italic: true,
			Size:   14, // Updated to 14
			Color:  "000000",
			Family: "Times New Roman",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
}

func createMetaStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Color:  "666666",
			Family: "Times New Roman",
			Italic: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
		},
	})
}

func createHeaderStyle(f *excelize.File) (int, error) {
	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12, // User requested 12
			Color:  "000000",
			Family: "Times New Roman",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"92D050"}, // Green like in usage
			Pattern: 1,
		},
	})

	if err != nil {
		return 0, fmt.Errorf("create style error: %w", err)
	}

	return style, nil
}

func createDataStyle(f *excelize.File) (int, error) {

	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   12, // User requested 12
			Family: "Times New Roman",
		},
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			Horizontal: "center", // User requested align center
			WrapText:   true,
		},
		Border: borderFull(), // User requested border
	})
}

func createDateStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		NumFmt: 22,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: borderLight(),
	})
}

func createDataDateStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   12, // User requested 12
			Family: "Times New Roman",
		},
		NumFmt: 22,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: borderFull(),
	})
}

func createEvenRowStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   12, // User requested 12
			Family: "Times New Roman",
		},
		Fill: excelize.Fill{
			Type: "pattern", Color: []string{"F7F7F7"}, Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			Horizontal: "center", // Align center
			WrapText:   true,
		},
		Border: borderFull(),
	})
}

func createEvenDateStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   12, // User requested 12
			Family: "Times New Roman",
		},
		NumFmt: 22,
		Fill: excelize.Fill{
			Type: "pattern", Color: []string{"F7F7F7"}, Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: borderFull(),
	})
}

// Border màu đen rõ hơn (không phải borderLight gray)
func borderFull() []excelize.Border {
	return []excelize.Border{
		{Type: "left", Style: 1, Color: "000000"},
		{Type: "right", Style: 1, Color: "000000"},
		{Type: "top", Style: 1, Color: "000000"},
		{Type: "bottom", Style: 1, Color: "000000"},
	}
}

func borderLight() []excelize.Border {
	return []excelize.Border{
		{Type: "left", Style: 1, Color: "D9D9D9"},
		{Type: "right", Style: 1, Color: "D9D9D9"},
		{Type: "top", Style: 1, Color: "D9D9D9"},
		{Type: "bottom", Style: 1, Color: "D9D9D9"},
	}
}

func applyAgeConditional(f *excelize.File, sheet string, cols []string) {
	ageCol := -1
	for i, c := range cols {
		if c == "age" {
			ageCol = i + 1
			break
		}
	}
	if ageCol == -1 {
		return
	}

	col, _ := excelize.ColumnNumberToName(ageCol)

	redStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"FFC7CE"},
			Pattern: 1,
		},
	})

	_ = f.SetConditionalFormat(
		sheet,
		col+"3:"+col+"1000000",
		[]excelize.ConditionalFormatOptions{
			{
				Type:     "cell",
				Criteria: ">",
				Value:    "30",
				Format:   &redStyle,
			},
		},
	)
}

func configureFreezePane(f *excelize.File, sheetName string, freezeRow int) error {
	topLeftCell := fmt.Sprintf("A%d", freezeRow+1)
	if err := f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      freezeRow,
		TopLeftCell: topLeftCell,
	}); err != nil {
		return fmt.Errorf("set freeze pane error: %w", err)
	}

	return nil
}
