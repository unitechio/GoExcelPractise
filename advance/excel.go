package exporter

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

	// ===== STYLES (Init all styles first) =====
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

	// ===== INIT STREAM WRITER (Optimization for >1M rows) =====
	sw, err := f.NewStreamWriter(sheet)
	if err != nil {
		return "", err
	}

	// Helper to create Cell with Style
	newCell := func(val interface{}, styleID int) excelize.Cell {
		return excelize.Cell{Value: val, StyleID: styleID}
	}

	// ===== 1. PREPARE TOP HEADER DATA (Row 1-6) =====
	// Mẹo: Với Stream, ta ghi dữ liệu vào ô đầu tiên của vùng Merge, các ô sau để trống

	splitColIndex := 3
	if len(cols) < 5 {
		splitColIndex = 2
	}
	// leftMergeTo, _ := excelize.ColumnNumberToName(splitColIndex)
	rightMergeFromCol := splitColIndex + 1
	// rightMergeFrom, _ := excelize.ColumnNumberToName(rightMergeFromCol)

	// Row 1
	row1 := make([]interface{}, len(cols))
	row1[0] = newCell("CÔNG TY CỔ PHẦN GIAO THÔNG SỐ VIỆT NAM", topBoldStyle) // A1
	if rightMergeFromCol <= len(cols) {
		row1[rightMergeFromCol-1] = newCell("CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM", topBoldStyle)
	}
	if err := sw.SetRow("A1", row1); err != nil {
		return "", err
	}

	// Row 2
	row2 := make([]interface{}, len(cols))
	row2[0] = newCell("PHÒNG KINH DOANH", topBoldStyle) // A2
	if rightMergeFromCol <= len(cols) {
		row2[rightMergeFromCol-1] = newCell("Độc lập - Tự do - Hạnh phúc", topBoldStyle)
	}
	if err := sw.SetRow("A2", row2); err != nil {
		return "", err
	}

	// Row 3
	row3 := make([]interface{}, len(cols))
	row3[0] = newCell("Số: ...../ĐN", topNormalStyle) // A3
	if rightMergeFromCol <= len(cols) {
		dateStr := fmt.Sprintf("Hà Nội, ngày %s tháng %s năm %s", time.Now().Format("02"), time.Now().Format("01"), time.Now().Format("2006"))
		row3[rightMergeFromCol-1] = newCell(dateStr, topItalicStyle)
	}
	if err := sw.SetRow("A3", row3); err != nil {
		return "", err
	}

	// Row 5: TITLE
	row5 := make([]interface{}, len(cols))
	row5[0] = newCell(reportTitle, titleStyle)
	if err := sw.SetRow("A5", row5); err != nil {
		return "", err
	}

	// ===== 2. TABLE HEADER (Row 7) =====
	headerRowIndex := 7
	headerRowData := make([]interface{}, len(cols))
	maxLen := make([]int, len(cols))

	for i, col := range cols {
		text := beautifyHeader(col)
		headerRowData[i] = newCell(text, headerStyle)
		maxLen[i] = utf8.RuneCountInString(text)
	}
	cellHeader, _ := excelize.CoordinatesToCellName(1, headerRowIndex)
	if err := sw.SetRow(cellHeader, headerRowData); err != nil {
		return "", err
	}

	// ===== 3. DATA ROWS (Row 8+) =====
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

		rowData := make([]interface{}, len(cols))

		// Determine row style
		currentRowStyle := evenRowStyle
		currentDateStyle := evenDateStyle
		if rowIndex%2 != 0 {
			currentRowStyle = dataStyle
			currentDateStyle = dateDateStyle
		}

		for i, v := range values {
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

			rowData[i] = newCell(val, style)

			// Track width (sample 200 rows)
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

		cellRef, _ := excelize.CoordinatesToCellName(1, rowIndex)
		if err := sw.SetRow(cellRef, rowData); err != nil {
			return "", err
		}
		rowIndex++
	}

	// ===== 4. FLUSH STREAM (IMPORTANT) =====
	if err := sw.Flush(); err != nil {
		return "", err
	}

	// ===== 5. POST-PROCESSING (MERGE & FORMATTING) =====
	// Apply Logic Layout (Merge Cells)
	leftMergeTo, _ := excelize.ColumnNumberToName(splitColIndex)
	rightMergeFrom, _ := excelize.ColumnNumberToName(rightMergeFromCol)

	// Top Header Merges (Must be done AFTER Flush)
	_ = f.MergeCell(sheet, "A1", leftMergeTo+"1")
	_ = f.MergeCell(sheet, rightMergeFrom+"1", lastCol+"1")
	_ = f.MergeCell(sheet, "A2", leftMergeTo+"2")
	_ = f.MergeCell(sheet, rightMergeFrom+"2", lastCol+"2")
	_ = f.MergeCell(sheet, "A3", leftMergeTo+"3")
	_ = f.MergeCell(sheet, rightMergeFrom+"3", lastCol+"3")

	// Report Title Merge
	_ = f.MergeCell(sheet, "A5", lastCol+"5")

	// Set Row Heights
	_ = f.SetRowHeight(sheet, 1, 26)              // Font size 14
	_ = f.SetRowHeight(sheet, 2, 26)              // Font size 14
	_ = f.SetRowHeight(sheet, 3, 26)              // Font size 14
	_ = f.SetRowHeight(sheet, 5, 30)              // Report Title
	_ = f.SetRowHeight(sheet, headerRowIndex, 30) // Header Table

	// Set Data Row Heights (Set Default for Sheet is best for 1M rows)
	// For huge datasets, looping SetRowHeight is slow. We set default.
	// But SetRowHeight has precedence over DefaultRowHeight.
	// Since we used StreamWriter, the rows exist.
	// Optimization for UI: Loop might be needed if DefaultRowHeight doesn't apply to existing rows.
	// However, for 1M rows, we should SKIP SetRowHeight loop to avoid timeout.
	// Let's rely on content or set a global property?
	// Excelize SetSheetProps DefaultRowHeight applies to new rows usually.
	// Let's try to set it. if it doesn't work, user won't have 24px height but performance is safe.
	defaultHeight := 24.0
	_ = f.SetSheetProps(sheet, &excelize.SheetPropsOptions{
		DefaultRowHeight: &defaultHeight,
	})

	// If the above doesn't work for existing stream rows, we can't easily loop 1M times.
	// Compromise: Loop for first 5000 rows to look nice, others default?
	// Or just accept standard height for data rows in exchange for performance.
	// Let's try loop up to a limit.
	limitHeightSet := 5000
	if rowIndex < limitHeightSet {
		limitHeightSet = rowIndex
	}
	for r := 8; r < limitHeightSet; r++ {
		_ = f.SetRowHeight(sheet, r, 24)
	}

	// Freeze Pane
	_ = configureFreezePane(f, sheet, 6)

	// Auto Filter
	filterRange := fmt.Sprintf("A%d:%s%d", headerRowIndex, lastCol, headerRowIndex)
	_ = f.AutoFilter(sheet, filterRange, nil)

	// Auto Width
	for i, l := range maxLen {
		col, _ := excelize.ColumnNumberToName(i + 1)
		// Improved width calculation: Font Times New Roman 12/14 is wide.
		// Factor 2.8 + padding 10 is safe for Dates and Long text.
		width := float64(l)*2.8 + 10

		// CRITICAL: Set MIN WIDTH to 20.
		// This ensures that merged cells (Header) have enough total width
		// even if data columns (ID, Age) are narrow.
		if width < 20 {
			width = 20
		}

		if width > 100 { // Allow wider max width
			width = 100
		}
		_ = f.SetColWidth(sheet, col, col, width)
	}

	// Logo
	_ = f.AddPicture(sheet, "A1", "./exporter/logo.png", &excelize.GraphicOptions{
		ScaleX: 0.3, ScaleY: 0.3, OffsetX: 10, OffsetY: 5,
	})

	if err := f.SaveAs(filePath); err != nil {
		return "", err
	}

	log.Printf("[XLSX] export done in %v", time.Since(start))
	return filePath, nil
}
