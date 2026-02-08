package basic

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
	"github.com/xuri/excelize/v2"
)

func CreateExcelStream(rows *sqlx.Rows, filePrefix string, reportTitle string) (string, error) {
	defer rows.Close()

	start := time.Now()

	fileName := fmt.Sprintf("%s_%d.xlsx", filePrefix, time.Now().Unix())
	filePath := filepath.Join(os.TempDir(), fileName)

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Report"
	f.SetSheetName("Sheet1", sheet)

	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	// ===== STYLES =====
	titleStyle, _ := createTitleStyle(f)
	// metaStyle, _ := createMetaStyle(f) // No longer used
	headerStyle, _ := createHeaderStyle(f)
	dataStyle, _ := createDataStyle(f)
	dateDateStyle, _ := createDataDateStyle(f)
	evenRowStyle, _ := createEvenRowStyle(f)
	evenDateStyle, _ := createEvenDateStyle(f)

	// New Styles for Top Header
	topBoldStyle, _ := createTopHeaderBoldStyle(f)
	topNormalStyle, _ := createTopHeaderNormalStyle(f)
	topItalicStyle, _ := createTopHeaderItalicStyle(f)

	lastCol, _ := excelize.ColumnNumberToName(len(cols))

	// Fixed width for header to ensure it's always readable (approx A4 width)
	// regardless of whether data has few or many columns.
	headerColCount := 8
	headerLastCol, _ := excelize.ColumnNumberToName(headerColCount)

	// ===== TOP HEADER (Row 1-3) =====
	// Left Side: Company Info (Merge A-C)
	// Right Side: Republic Info (Merge D-H)

	splitColIndex := 3
	// Logic splitColIndex remains effective if we ever change headerColCount
	if headerColCount < 5 {
		splitColIndex = 2
	}
	leftMergeTo, _ := excelize.ColumnNumberToName(splitColIndex)
	rightMergeFrom, _ := excelize.ColumnNumberToName(splitColIndex + 1)

	// Row 1
	f.MergeCell(sheet, "A1", leftMergeTo+"1")
	f.SetCellValue(sheet, "A1", "CÔNG TY CỔ PHẦN GIAO THÔNG SỐ VIỆT NAM")
	f.SetCellStyle(sheet, "A1", "A1", topBoldStyle)

	f.MergeCell(sheet, rightMergeFrom+"1", headerLastCol+"1")
	f.SetCellValue(sheet, rightMergeFrom+"1", "CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM")
	f.SetCellStyle(sheet, rightMergeFrom+"1", rightMergeFrom+"1", topBoldStyle)

	f.SetRowHeight(sheet, 1, 26)

	// Row 2
	f.MergeCell(sheet, "A2", leftMergeTo+"2")
	f.SetCellValue(sheet, "A2", "PHÒNG KINH DOANH")
	f.SetCellStyle(sheet, "A2", "A2", topBoldStyle)

	f.MergeCell(sheet, rightMergeFrom+"2", headerLastCol+"2")
	f.SetCellValue(sheet, rightMergeFrom+"2", "Độc lập - Tự do - Hạnh phúc")
	f.SetCellStyle(sheet, rightMergeFrom+"2", rightMergeFrom+"2", topBoldStyle)

	f.SetRowHeight(sheet, 2, 26)

	// Row 3 (Date info)
	f.MergeCell(sheet, "A3", leftMergeTo+"3")
	f.SetCellValue(sheet, "A3", "Số: ...../ĐN") // Placeholder
	f.SetCellStyle(sheet, "A3", "A3", topNormalStyle)

	f.MergeCell(sheet, rightMergeFrom+"3", headerLastCol+"3")
	f.SetCellValue(sheet, rightMergeFrom+"3", fmt.Sprintf("Hà Nội, ngày %s tháng %s năm %s",
		time.Now().Format("02"), time.Now().Format("01"), time.Now().Format("2006")))
	f.SetCellStyle(sheet, rightMergeFrom+"3", rightMergeFrom+"3", topItalicStyle)

	f.SetRowHeight(sheet, 3, 26)

	// Row 5: REPORT TITLE
	f.MergeCell(sheet, "A5", headerLastCol+"5")
	f.SetCellValue(sheet, "A5", reportTitle) // BÁO CÁO DANH SÁCH...
	f.SetCellStyle(sheet, "A5", headerLastCol+"5", titleStyle)
	f.SetRowHeight(sheet, 5, 30)

	// ===== FREEZE PANE (From Row 7) =====
	_ = configureFreezePane(f, sheet, 7) // Freeze top 6 rows

	// ===== HEADER ROW (Row 7) =====
	headerRowIndex := 7
	maxLen := make([]int, len(cols))
	for i, col := range cols {
		text := beautifyHeader(col)
		cellName, _ := excelize.CoordinatesToCellName(i+1, headerRowIndex)
		f.SetCellValue(sheet, cellName, text)
		f.SetCellStyle(sheet, cellName, cellName, headerStyle)
		maxLen[i] = utf8.RuneCountInString(text)
	}
	f.SetRowHeight(sheet, headerRowIndex, 30)

	// ===== AUTOFILTER =====
	filterRange := fmt.Sprintf("A%d:%s%d", headerRowIndex, lastCol, headerRowIndex)
	_ = f.AutoFilter(sheet, filterRange, nil)

	// ===== DATA (Row 8+) using Normal API =====
	rowIndex := 8
	values := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return "", err
		}

		// Apply alternating row style
		currentRowStyle := evenRowStyle
		currentDateStyle := evenDateStyle
		if rowIndex%2 != 0 {
			currentRowStyle = dataStyle
			currentDateStyle = dateDateStyle
		}

		for i, v := range values {
			cellName, _ := excelize.CoordinatesToCellName(i+1, rowIndex)

			var val interface{}
			style := currentRowStyle
			isDate := false

			switch t := v.(type) {
			case nil:
				val = ""
			case time.Time:
				val = t
				style = currentDateStyle
				isDate = true
			case []byte:
				val = softWrapEmail(string(t))
			default:
				val = softWrapEmail(fmt.Sprintf("%v", t))
			}

			// Use Normal API for 100% correct formatting
			f.SetCellValue(sheet, cellName, val)
			f.SetCellStyle(sheet, cellName, cellName, style)

			// Track max length
			if rowIndex <= 200 {
				var displayLen int
				if isDate {
					displayLen = 19
				} else {
					displayLen = utf8.RuneCountInString(fmt.Sprintf("%v", val))
				}
				if displayLen > maxLen[i] {
					maxLen[i] = displayLen
				}
			}
		}

		// Set row height for better UI
		f.SetRowHeight(sheet, rowIndex, 24)

		rowIndex++
	}

	// ===== AUTO WIDTH =====
	for i, l := range maxLen {
		col, _ := excelize.ColumnNumberToName(i + 1)
		// Improved width calculation: more generous spacing
		width := float64(l)*1.8 + 6
		if width < 16 {
			width = 16
		}
		if width > 50 {
			width = 50
		}
		_ = f.SetColWidth(sheet, col, col, width)
	}

	// ===== CONDITIONAL FORMAT (AGE) =====
	// applyAgeConditional(f, sheet, cols) // Disabled to avoid conflict

	if err := f.SaveAs(filePath); err != nil {
		return "", err
	}

	log.Printf("[XLSX] export done in %v", time.Since(start))
	return filePath, nil
}
