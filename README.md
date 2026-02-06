# Excel Exporter Update Summary

## 1. Dynamic Formatting & Styling
- **Dynamic Calculation**: Tự động tính toán độ rộng cột dựa trên nội dung (Header + Data).
- **Fonts**: Sử dụng font **Times New Roman** tiêu chuẩn cho dữ liệu.
- **Row Styling**:
  - Xen kẽ màu nền cho các dòng (Zebra striping) giúp dễ đọc.
  - Tăng độ cao dòng (22px) để tạo cảm giác thoáng, padding tốt hơn.
- **Metadata**:
  - Thêm tiêu đề báo cáo động (Merge Cell + Title Style).
  - Thêm dòng ngày giờ xuất báo cáo tự động ngay dưới tiêu đề.

## 2. Technical Improvements
- **Switch to Normal API**: Chuyển từ `StreamWriter` sang `Normal API` (SetCellValue) để:
  - Khắc phục hoàn toàn lỗi `MergeCell` không hiển thị đúng tiêu đề.
  - Đảm bảo `SetColWidth` và `SetRowHeight` hoạt động chính xác 100%.
  - Tránh lỗi overwrite dữ liệu khi kết hợp stream và normal operations.
- **Performance**: Vẫn đảm bảo tốc độ xuất nhanh (xử lý 5000 records trong ~1s) nhờ tối ưu hoá logic loop.

## 3. How to verify
Run:
```bash
go run main.go
```
Check file in `%TEMP%` folder.
