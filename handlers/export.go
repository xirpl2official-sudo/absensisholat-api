package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"absensholat-api/models"
)

type ExportAbsensiRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Kelas     string `form:"kelas"`
	Jurusan   string `form:"jurusan"`
}

// Header constants
const (
	CSVContentType   = "text/csv"
	ExcelContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
)

var (
	// Headers for CSV exports
	csvAbsensiHeaders = []string{"No", "NIS", "Nama Siswa", "Kelas", "Jurusan", "Tanggal", "Status", "Deskripsi"}
	csvLaporanHeaders = []string{"Status", "Jumlah", "Persentase"}
	csvDetailHeaders  = csvAbsensiHeaders // Same structure for detail section
)

// Helper function to write CSV rows with error handling
func writeCSVRow(writer *csv.Writer, logger *zap.SugaredLogger, row []string, context string) error {
	if err := writer.Write(row); err != nil {
		logger.Errorw("Failed to write CSV row",
			"error", err.Error(),
			"context", context,
			"row_data", row, // Include row data for debugging
		)
		return fmt.Errorf("failed to write CSV %s row: %w", context, err)
	}
	return nil
}

// Helper function to write multiple CSV rows with error handling
func writeCSVRows(writer *csv.Writer, logger *zap.SugaredLogger, rows [][]string, context string) error {
	for _, row := range rows {
		if err := writeCSVRow(writer, logger, row, context); err != nil {
			return err
		}
	}
	return nil
}

// ExportAbsensiCSV godoc
// @Summary Export data absensi ke CSV
// @Description Mengexport data absensi siswa dalam format CSV. Dapat difilter berdasarkan tanggal, kelas, dan jurusan
// @Tags export
// @Accept json
// @Produce text/csv
// @Param start_date query string false "Tanggal mulai (format: YYYY-MM-DD)"
// @Param end_date query string false "Tanggal akhir (format: YYYY-MM-DD)"
// @Param kelas query string false "Filter berdasarkan kelas"
// @Param jurusan query string false "Filter berdasarkan jurusan"
// @Security BearerAuth
// @Success 200 {file} file "File CSV berhasil didownload"
// @Failure 401 {object} ErrorResponse "Tidak terotentikasi"
// @Failure 40 ErrorResponse "Akses ditolak"
// @Failure 500 {object} ErrorResponse "Kesalahan server internal"
// @Router /export/absensi [get]
func ExportAbsensiCSV(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ExportAbsensiRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			logger.Warnw("Invalid export request",
				"error", err.Error(),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Parameter tidak valid",
				"error":   err.Error(),
			})
			return
		}

		// Build query
		query := db.Model(&models.Absensi{}).
			Preload("Siswa").
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis")

		// Apply date filters
		if req.StartDate != "" {
			query = query.Where("DATE(absensi.tanggal) >= ?", req.StartDate)
		}
		if req.EndDate != "" {
			query = query.Where("DATE(absensi.tanggal) <= ?", req.EndDate)
		}

		// Apply class and department filters
		if req.Kelas != "" {
			query = query.Where("siswa.kelas = ?", req.Kelas)
		}
		if req.Jurusan != "" {
			query = query.Where("siswa.jurusan = ?", req.Jurusan)
		}

		var absensiList []models.Absensi
		if err := query.Order("absensi.tanggal DESC, siswa.kelas, siswa.nis").Find(&absensiList).Error; err != nil {
			logger.Errorw("Failed to fetch absensi for export",
				"error", err.Error(),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal mengambil data absensi",
			})
			return
		}

		// Set CSV headers
		filename := fmt.Sprintf("absensi_export_%s.csv", time.Now().Format("20060102_150405"))
		c.Header("Content-Type", CSVContentType)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush() // Ensure buffer is flushed

		// Write CSV header
		if err := writeCSVRow(writer, logger, csvAbsensiHeaders, "header"); err != nil {
			// Error already logged by writeCSVRow
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal menulis header CSV",
				"error":   err.Error(),
			})
			return
		}

		// Write data rows
		for i, absensi := range absensiList {
			namaSiswa := ""
			kelas := ""
			jurusan := ""
			if absensi.Siswa != nil {
				namaSiswa = absensi.Siswa.NamaSiswa
				kelas = absensi.Siswa.Kelas
				jurusan = absensi.Siswa.Jurusan
			}

			row := []string{
				fmt.Sprintf("%d", i+1),
				absensi.NIS,
				namaSiswa,
				kelas,
				jurusan,
				absensi.Tanggal.Format("2006-01-02"),
				absensi.Status,
				absensi.Deskripsi,
			}
			if err := writeCSVRow(writer, logger, row, fmt.Sprintf("data_row_%d", i+1)); err != nil {
				// Error already logged by writeCSVRow
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Gagal menulis data CSV",
					"error":   err.Error(),
				})
				return
			}
		}

		logger.Infow("Absensi exported to CSV successfully",
			"total_records", len(absensiList),
			"start_date", req.StartDate,
			"end_date", req.EndDate,
		)
	}
}

type LaporanStatistik struct {
	TotalSiswa        int64   `json:"total_siswa"`
	TotalAbsensi      int64   `json:"total_absensi"`
	TotalHadir        int64   `json:"total_hadir"`
	TotalIzin         int64   `json:"total_izin"`
	TotalSakit        int64   `json:"total_sakit"`
	TotalAlpha        int64   `json:"total_alpha"`
	PersentaseHadir   float64 `json:"persentase_hadir"`
	PersentaseIzin    float64 `json:"persentase_izin"`
	PersentaseSakit   float64 `json:"persentase_sakit"`
	PersentaseAlpha   float64 `json:"persentase_alpha"`
	RataRataKehadiran float64 `json:"rata_rata_kehadiran"`
}

// ExportLaporanCSV godoc
// @Summary Export laporan absensi dengan statistik ke CSV
// @Description Mengexport laporan absensi dengan statistik kehadiran dalam format CSV. Termasuk persentase dan rata-rata kehadiran
// @Tags export
// @Accept json
// @Produce text/csv
// @Param start_date query string false "Tanggal mulai (format: YYYY-MM-DD)"
// @Param end_date query string false "Tanggal akhir (format: YYYY-MM-DD)"
// @Param kelas query string false "Filter berdasarkan kelas"
// @Param jurusan query string false "Filter berdasarkan jurusan"
// @Security BearerAuth
// @Success 200 {file} file "File CSV laporan berhasil didownload"
// @Failure 401 {object} ErrorResponse "Tidak terotentikasi"
// @Failure 403 {object} ErrorResponse "Akses ditolak"
// @Failure 500 {object} ErrorResponse "Kesalahan server internal"
// @Router /export/laporan [get]
func ExportLaporanCSV(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ExportAbsensiRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			logger.Warnw("Invalid export request",
				"error", err.Error(),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Parameter tidak valid",
				"error":   err.Error(),
			})
			return
		}

		// Build base query for statistics
		baseQuery := db.Model(&models.Absensi{}).
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis")

		if req.StartDate != "" {
			baseQuery = baseQuery.Where("DATE(absensi.tanggal) >= ?", req.StartDate)
		}
		if req.EndDate != "" {
			baseQuery = baseQuery.Where("DATE(absensi.tanggal) <= ?", req.EndDate)
		}
		if req.Kelas != "" {
			baseQuery = baseQuery.Where("siswa.kelas = ?", req.Kelas)
		}
		if req.Jurusan != "" {
			baseQuery = baseQuery.Where("siswa.jurusan = ?", req.Jurusan)
		}

		// Get statistics
		var stats LaporanStatistik

		// Total siswa in filter
		siswaQuery := db.Model(&models.Siswa{})
		if req.Kelas != "" {
			siswaQuery = siswaQuery.Where("kelas = ?", req.Kelas)
		}
		if req.Jurusan != "" {
			siswaQuery = siswaQuery.Where("jurusan = ?", req.Jurusan)
		}
		siswaQuery.Count(&stats.TotalSiswa)

		// Count by status
		baseQuery.Count(&stats.TotalAbsensi)

		db.Model(&models.Absensi{}).
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis").
			Where(buildWhereClause(req)).
			Where("status = ?", "hadir").
			Count(&stats.TotalHadir)

		db.Model(&models.Absensi{}).
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis").
			Where(buildWhereClause(req)).
			Where("status = ?", "izin").
			Count(&stats.TotalIzin)

		db.Model(&models.Absensi{}).
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis").
			Where(buildWhereClause(req)).
			Where("status = ?", "sakit").
			Count(&stats.TotalSakit)

		db.Model(&models.Absensi{}).
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis").
			Where(buildWhereClause(req)).
			Where("status = ?", "alpha").
			Count(&stats.TotalAlpha)

		// Calculate percentages
		if stats.TotalAbsensi > 0 {
			stats.PersentaseHadir = float64(stats.TotalHadir) / float64(stats.TotalAbsensi) * 100
			stats.PersentaseIzin = float64(stats.TotalIzin) / float64(stats.TotalAbsensi) * 100
			stats.PersentaseSakit = float64(stats.TotalSakit) / float64(stats.TotalAbsensi) * 100
			stats.PersentaseAlpha = float64(stats.TotalAlpha) / float64(stats.TotalAbsensi) * 100
		}

		// Calculate average attendance per student
		if stats.TotalSiswa > 0 && stats.TotalAbsensi > 0 {
			stats.RataRataKehadiran = float64(stats.TotalHadir) / float64(stats.TotalAbsensi) * 100
		}

		// Get detailed absensi data
		var absensiList []models.Absensi
		detailQuery := db.Model(&models.Absensi{}).
			Preload("Siswa").
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis")

		if req.StartDate != "" {
			detailQuery = detailQuery.Where("DATE(absensi.tanggal) >= ?", req.StartDate)
		}
		if req.EndDate != "" {
			detailQuery = detailQuery.Where("DATE(absensi.tanggal) <= ?", req.EndDate)
		}
		if req.Kelas != "" {
			detailQuery = detailQuery.Where("siswa.kelas = ?", req.Kelas)
		}
		if req.Jurusan != "" {
			detailQuery = detailQuery.Where("siswa.jurusan = ?", req.Jurusan)
		}

		if err := detailQuery.Order("absensi.tanggal DESC, siswa.kelas, siswa.nis").Find(&absensiList).Error; err != nil {
			logger.Errorw("Failed to fetch absensi for laporan",
				"error", err.Error(),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal mengambil data absensi",
			})
			return
		}

		// Set CSV headers
		filename := fmt.Sprintf("laporan_absensi_%s.csv", time.Now().Format("20060102_150405"))
		c.Header("Content-Type", CSVContentType)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Prepare and write sections
		// 1. Statistics Summary Section
		statSectionTitle := []string{"LAPORAN STATISTIK ABSENSI SHOLAT"}
		emptyRow := []string{""}

		// Filters
		var filterRows [][]string
		if req.StartDate != "" || req.EndDate != "" {
			periode := "Semua Waktu"
			if req.StartDate != "" && req.EndDate != "" {
				periode = fmt.Sprintf("%s s/d %s", req.StartDate, req.EndDate)
			} else if req.StartDate != "" {
				periode = fmt.Sprintf("Dari %s", req.StartDate)
			} else if req.EndDate != "" {
				periode = fmt.Sprintf("Sampai %s", req.EndDate)
			}
			filterRows = append(filterRows, []string{"Periode", periode})
		}
		if req.Kelas != "" {
			filterRows = append(filterRows, []string{"Kelas", req.Kelas})
		}
		if req.Jurusan != "" {
			filterRows = append(filterRows, []string{"Jurusan", req.Jurusan})
		}

		// Summary Stats
		summaryRows := [][]string{
			{"RINGKASAN STATISTIK"},
			{"Total Siswa", fmt.Sprintf("%d", stats.TotalSiswa)},
			{"Total Absensi", fmt.Sprintf("%d", stats.TotalAbsensi)},
			emptyRow,          // Empty row before status breakdown
			csvLaporanHeaders, // Header for status table
			{"Hadir", fmt.Sprintf("%d", stats.TotalHadir), fmt.Sprintf("%.2f%%", stats.PersentaseHadir)},
			{"Izin", fmt.Sprintf("%d", stats.TotalIzin), fmt.Sprintf("%.2f%%", stats.PersentaseIzin)},
			{"Sakit", fmt.Sprintf("%d", stats.TotalSakit), fmt.Sprintf("%.2f%%", stats.PersentaseSakit)},
			{"Alpha", fmt.Sprintf("%d", stats.TotalAlpha), fmt.Sprintf("%.2f%%", stats.PersentaseAlpha)},
			emptyRow,           // Empty row before{"Rata-rata Kehadiran", fmt.Sprintf("%.2f%%", stats.RataRataKehadiran)},
			emptyRow, emptyRow, // Two empty rows before detail section
			{"DETAIL ABSENSI"},
			csvDetailHeaders, // Header for detail table
		}

		// Combine all rows for the statistics section
		allStatRows := [][]string{statSectionTitle, emptyRow}
		allStatRows = append(allStatRows, filterRows...)
		allStatRows = append(allStatRows, emptyRow) // Empty row after filters
		allStatRows = append(allStatRows, summaryRows...)

		if err := writeCSVRows(writer, logger, allStatRows, "statistics_section"); err != nil {
			// Error already logged by writeCSVRows/writeCSVRow
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal menulis bagian statistik CSV",
				"error":   err.Error(),
			})
			return
		}

		// 2. Detail Data Rows Section
		for i, absensi := range absensiList {
			namaSiswa := ""
			kelas := ""
			jurusan := ""
			if absensi.Siswa != nil {
				namaSiswa = absensi.Siswa.NamaSiswa
				kelas = absensi.Siswa.Kelas
				jurusan = absensi.Siswa.Jurusan
			}

			row := []string{
				fmt.Sprintf("%d", i+1),
				absensi.NIS,
				namaSiswa,
				kelas,
				jurusan,
				absensi.Tanggal.Format("2006-01-02"),
				absensi.Status,
				absensi.Deskripsi,
			}
			if err := writeCSVRow(writer, logger, row, fmt.Sprintf("detail_row_%d", i+1)); err != nil {
				// Error already logged by writeCSVRow
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Gagal menulis bagian detail CSV",
					"error":   err.Error(),
				})
				return
			}
		}

		logger.Infow("Laporan exported to CSV successfully",
			"total_records", len(absensiList),
			"stats", stats,
		)
	}
}

func buildWhereClause(req ExportAbsensiRequest) string {
	where := "1=1"
	if req.StartDate != "" {
		where += fmt.Sprintf(" AND DATE(absensi.tanggal) >= '%s'", req.StartDate)
	}
	if req.EndDate != "" {
		where += fmt.Sprintf(" AND DATE(absensi.tanggal) <= '%s'", req.EndDate)
	}
	if req.Kelas != "" {
		where += fmt.Sprintf(" AND siswa.kelas = '%s'", req.Kelas)
	}
	if req.Jurusan != "" {
		where += fmt.Sprintf(" AND siswa.jurusan = '%s'", req.Jurusan)
	}
	return where
}

// ExportAbsensiExcel godoc
// @Summary Export data absensi ke Excel
// @Description Mengexport data absensi siswa dalam format Excel. Dapat difilter berdasarkan tanggal, kelas, dan jurusan
// @Tags export
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param start_date query string false "Tanggal mulai (format: YYYY-MM-DD)"
// @Param end_date query string false "Tanggal akhir (format: YYYY-MM-DD)"
// @Param kelas query string false "Filter berdasarkan kelas"
// @Param jurusan query string false "Filter berdasarkan jurusan"
// @Security BearerAuth
// @Success 200 {file} file "File Excel berhasil didownload"
// @Failure 401 {object} ErrorResponse "Tidak terotentikasi"
// @Failure 403 {object} ErrorResponse "Akses ditolak"
// @Failure 500 {object} ErrorResponse "Kesalahan server internal"
// @Router /export/absensi/excel [get]
func ExportAbsensiExcel(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ExportAbsensiRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			logger.Warnw("Invalid export request",
				"error", err.Error(),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Parameter tidak valid",
				"error":   err.Error(),
			})
			return
		}

		// Build query
		query := db.Model(&models.Absensi{}).
			Preload("Siswa").
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis")

		// Apply date filters
		if req.StartDate != "" {
			query = query.Where("DATE(absensi.tanggal) >= ?", req.StartDate)
		}
		if req.EndDate != "" {
			query = query.Where("DATE(absensi.tanggal) <= ?", req.EndDate)
		}

		// Apply class and department filters
		if req.Kelas != "" {
			query = query.Where("siswa.kelas = ?", req.Kelas)
		}
		if req.Jurusan != "" {
			query = query.Where("siswa.jurusan = ?", req.Jurusan)
		}

		var absensiList []models.Absensi
		if err := query.Order("absensi.tanggal DESC, siswa.kelas, siswa.nis").Find(&absensiList).Error; err != nil {
			logger.Errorw("Failed to fetch absensi for export",
				"error", err.Error(),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal mengambil data absensi",
			})
			return
		}

		// Create Excel file
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				logger.Errorw("Failed to close Excel file",
					"error", err.Error(),
				)
			}
		}()

		// Set sheet name
		sheetName := "Absensi"
		if err := f.SetSheetName("Sheet1", sheetName); err != nil {
			logger.Errorw("Failed to set sheet name", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal menyiapkan file Excel",
				"error":   err.Error(),
			})
			return
		}

		setCellValue := func(sheet, cell string, value interface{}) error {
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				logger.Errorw("Failed to set cell value", "error", err.Error(), "cell", cell)
				return err
			}
			return nil
		}

		// Define styles
		headerStyle, err := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		})
		if err != nil {
			logger.Errorw("Failed to create header style", "error", err.Error())
			// Continue without style, but log the error
		}

		// Write header with styling
		headers := []string{"No", "NIS", "Nama Siswa", "Kelas", "Jurusan", "Tanggal", "Status", "Deskripsi"}
		for colIndex, header := range headers {
			cellName, err := excelize.CoordinatesToCellName(colIndex+1, 1) // Row 1
			if err != nil {
				logger.Errorw("Failed to calculate cell name for header", "error", err.Error(), "column", colIndex+1)
				// Consider aborting if this is critical
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Gagal menyiapkan file Excel",
					"error":   "Internal error calculating cell coordinates",
				})
				return
			}
			if err := setCellValue(sheetName, cellName, header); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Gagal menulis file Excel",
				})
				return
			}
			if headerStyle != 0 { // Only apply style if it was created successfully
				if err := f.SetCellStyle(sheetName, cellName, cellName, headerStyle); err != nil {
					logger.Warnw("Failed to set cell style", "error", err.Error())
				}
			}
		}

		// Write data rows
		for rowIndex, absensi := range absensiList {
			rowNum := rowIndex + 2 // Start from row 2 (after header)
			namaSiswa := ""
			kelas := ""
			jurusan := ""
			if absensi.Siswa != nil {
				namaSiswa = absensi.Siswa.NamaSiswa
				kelas = absensi.Siswa.Kelas
				jurusan = absensi.Siswa.Jurusan
			}

			// Calculate cell names for this row
			if err := setCellValue(sheetName, fmt.Sprintf("A%d", rowNum), rowIndex+1); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("B%d", rowNum), absensi.NIS); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("C%d", rowNum), namaSiswa); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("D%d", rowNum), kelas); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("E%d", rowNum), jurusan); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("F%d", rowNum), absensi.Tanggal.Format("2006-01-02")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("G%d", rowNum), absensi.Status); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("H%d", rowNum), absensi.Deskripsi); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
		}

		// Auto-adjust column widths
		colLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
		widths := []float64{5, 12, 20, 12, 15, 12, 12, 20}
		for i, colLetter := range colLetters {
			if err := f.SetColWidth(sheetName, colLetter, colLetter, widths[i]); err != nil {
				logger.Warnw("Failed to set column width", "error", err.Error(), "col", colLetter)
			}
		}

		// Set response headers
		filename := fmt.Sprintf("absensi_export_%s.xlsx", time.Now().Format("20060102_150405"))
		c.Header("Content-Type", ExcelContentType)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		// Write to response
		if err := f.Write(c.Writer); err != nil {
			logger.Errorw("Failed to write Excel file to response",
				"error", err.Error(),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal mengirim file Excel",
				"error":   err.Error(),
			})
			return
		}

		logger.Infow("Absensi exported to Excel successfully",
			"total_records", len(absensiList),
			"start_date", req.StartDate,
			"end_date", req.EndDate,
		)
	}
}

// ExportLaporanExcel godoc
// @Summary Export laporan absensi dengan statistik ke Excel
// @Description Mengexport laporan absensi dengan statistik kehadiran dalam format Excel. Termasuk persentase dan rata-rata kehadiran
// @Tags export
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param start_date query string false "Tanggal mulai (format: YYYY-MM-DD)"
// @Param end_date query string false "Tanggal akhir (format: YYYY-MM-DD)"
// @Param kelas query string false "Filter berdasarkan kelas"
// @Param jurusan query string false "Filter berdasarkan jurusan"
// @Security BearerAuth
// @Success 200 {file} file "File Excel laporan berhasil didownload"
// @Failure 401 {object} ErrorResponse "Tidak terotentikasi"
// @Failure 403 {object} ErrorResponse "Akses ditolak"
// @Failure 500 {object} ErrorResponse "Kesalahan server internal"
// @Router /export/laporan/excel [get]
func ExportLaporanExcel(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ExportAbsensiRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			logger.Warnw("Invalid export request",
				"error", err.Error(),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Parameter tidak valid",
				"error":   err.Error(),
			})
			return
		}

		// Build base query for statistics
		baseQuery := db.Model(&models.Absensi{}).
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis")

		if req.StartDate != "" {
			baseQuery = baseQuery.Where("DATE(absensi.tanggal) >= ?", req.StartDate)
		}
		if req.EndDate != "" {
			baseQuery = baseQuery.Where("DATE(absensi.tanggal) <= ?", req.EndDate)
		}
		if req.Kelas != "" {
			baseQuery = baseQuery.Where("siswa.kelas = ?", req.Kelas)
		}
		if req.Jurusan != "" {
			baseQuery = baseQuery.Where("siswa.jurusan = ?", req.Jurusan)
		}

		// Get statistics
		var stats LaporanStatistik

		// Total siswa in filter
		siswaQuery := db.Model(&models.Siswa{})
		if req.Kelas != "" {
			siswaQuery = siswaQuery.Where("kelas = ?", req.Kelas)
		}
		if req.Jurusan != "" {
			siswaQuery = siswaQuery.Where("jurusan = ?", req.Jurusan)
		}
		siswaQuery.Count(&stats.TotalSiswa)

		// Count by status
		baseQuery.Count(&stats.TotalAbsensi)

		db.Model(&models.Absensi{}).
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis").
			Where(buildWhereClause(req)).
			Where("status = ?", "hadir").
			Count(&stats.TotalHadir)

		db.Model(&models.Absensi{}).
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis").
			Where(buildWhereClause(req)).
			Where("status = ?", "izin").
			Count(&stats.TotalIzin)

		db.Model(&models.Absensi{}).
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis").
			Where(buildWhereClause(req)).
			Where("status = ?", "sakit").
			Count(&stats.TotalSakit)

		db.Model(&models.Absensi{}).
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis").
			Where(buildWhereClause(req)).
			Where("status = ?", "alpha").
			Count(&stats.TotalAlpha)

		// Calculate percentages
		if stats.TotalAbsensi > 0 {
			stats.PersentaseHadir = float64(stats.TotalHadir) / float64(stats.TotalAbsensi) * 100
			stats.PersentaseIzin = float64(stats.TotalIzin) / float64(stats.TotalAbsensi) * 100
			stats.PersentaseSakit = float64(stats.TotalSakit) / float64(stats.TotalAbsensi) * 100
			stats.PersentaseAlpha = float64(stats.TotalAlpha) / float64(stats.TotalAbsensi) * 100
		}

		// Calculate average attendance per student
		if stats.TotalSiswa > 0 && stats.TotalAbsensi > 0 {
			stats.RataRataKehadiran = float64(stats.TotalHadir) / float64(stats.TotalAbsensi) * 100
		}

		// Get detailed absensi data
		var absensiList []models.Absensi
		detailQuery := db.Model(&models.Absensi{}).
			Preload("Siswa").
			Joins("LEFT JOIN siswa ON absensi.nis = siswa.nis")

		if req.StartDate != "" {
			detailQuery = detailQuery.Where("DATE(absensi.tanggal) >= ?", req.StartDate)
		}
		if req.EndDate != "" {
			detailQuery = detailQuery.Where("DATE(absensi.tanggal) <= ?", req.EndDate)
		}
		if req.Kelas != "" {
			detailQuery = detailQuery.Where("siswa.kelas = ?", req.Kelas)
		}
		if req.Jurusan != "" {
			detailQuery = detailQuery.Where("siswa.jurusan = ?", req.Jurusan)
		}

		if err := detailQuery.Order("absensi.tanggal DESC, siswa.kelas, siswa.nis").Find(&absensiList).Error; err != nil {
			logger.Errorw("Failed to fetch absensi for laporan",
				"error", err.Error(),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal mengambil data absensi",
			})
			return
		}

		// Create Excel file
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				logger.Errorw("Failed to close Excel file",
					"error", err.Error(),
				)
			}
		}()

		// Rename default sheet
		sheetName := "Laporan"
		if err := f.SetSheetName("Sheet1", sheetName); err != nil {
			logger.Errorw("Failed to set sheet name", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal menyiapkan file Excel",
				"error":   err.Error(),
			})
			return
		}

		setCellValue := func(sheet, cell string, value interface{}) error {
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				logger.Errorw("Failed to set cell value", "error", err.Error(), "cell", cell)
				return err
			}
			return nil
		}

		// Define styles
		titleStyle, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true, Size: 14},
		})
		if err != nil {
			logger.Errorw("Failed to create title style", "error", err.Error())
		}

		subHeaderStyle, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		if err != nil {
			logger.Errorw("Failed to create sub header style", "error", err.Error())
		}

		tableHeaderStyle, err := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		})
		if err != nil {
			logger.Errorw("Failed to create table header style", "error", err.Error())
		}

		// Write title
		if err := setCellValue(sheetName, "A1", "LAPORAN STATISTIK ABSENSI SHOLAT"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyiapkan file Excel"})
			return
		}
		if titleStyle != 0 {
			if err := f.SetCellStyle(sheetName, "A1", "A1", titleStyle); err != nil {
				logger.Warnw("Failed to set cell style", "error", err.Error())
			}
		}

		row := 3

		// Write filters
		if req.StartDate != "" || req.EndDate != "" {
			periode := "Semua Waktu"
			if req.StartDate != "" && req.EndDate != "" {
				periode = fmt.Sprintf("%s s/d %s", req.StartDate, req.EndDate)
			} else if req.StartDate != "" {
				periode = fmt.Sprintf("Dari %s", req.StartDate)
			} else if req.EndDate != "" {
				periode = fmt.Sprintf("Sampai %s", req.EndDate)
			}
			if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), "Periode"); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyiapkan file Excel"})
				return
			}
			if subHeaderStyle != 0 {
				if err := f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), subHeaderStyle); err != nil {
					logger.Warnw("Failed to set style", "error", err.Error())
				}
			}
			if err := setCellValue(sheetName, fmt.Sprintf("B%d", row), periode); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyiapkan file Excel"})
				return
			}
			row++
		}

		if req.Kelas != "" {
			if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), "Kelas"); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyiapkan file Excel"})
				return
			}
			if subHeaderStyle != 0 {
				if err := f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), subHeaderStyle); err != nil {
					logger.Warnw("Failed to set style", "error", err.Error())
				}
			}
			if err := setCellValue(sheetName, fmt.Sprintf("B%d", row), req.Kelas); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyiapkan file Excel"})
				return
			}
			row++
		}

		if req.Jurusan != "" {
			if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), "Jurusan"); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyiapkan file Excel"})
				return
			}
			if subHeaderStyle != 0 {
				if err := f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), subHeaderStyle); err != nil {
					logger.Warnw("Failed to set style", "error", err.Error())
				}
			}
			if err := setCellValue(sheetName, fmt.Sprintf("B%d", row), req.Jurusan); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyiapkan file Excel"})
				return
			}
			row++
		}

		row++ // Blank row after filters

		// Write statistics summary
		if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), "RINGKASAN STATISTIK"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		if subHeaderStyle != 0 {
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), subHeaderStyle); err != nil {
				logger.Warnw("Failed to set style", "error", err.Error())
			}
		}
		row++

		if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), "Total Siswa"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		if err := setCellValue(sheetName, fmt.Sprintf("B%d", row), stats.TotalSiswa); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		row++

		if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), "Total Absensi"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		if err := setCellValue(sheetName, fmt.Sprintf("B%d", row), stats.TotalAbsensi); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		row += 2 // Blank row before status table

		// Write status breakdown table header
		if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), "Status"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		if tableHeaderStyle != 0 {
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), tableHeaderStyle); err != nil {
				logger.Warnw("Failed to set style", "error", err.Error())
			}
		}
		if err := setCellValue(sheetName, fmt.Sprintf("B%d", row), "Jumlah"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		if err := setCellValue(sheetName, fmt.Sprintf("C%d", row), "Persentase"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		row++

		// Write status breakdown table data
		statusData := [][]interface{}{
			{"Hadir", stats.TotalHadir, fmt.Sprintf("%.2f%%", stats.PersentaseHadir)},
			{"Izin", stats.TotalIzin, fmt.Sprintf("%.2f%%", stats.PersentaseIzin)},
			{"Sakit", stats.TotalSakit, fmt.Sprintf("%.2f%%", stats.PersentaseSakit)},
			{"Alpha", stats.TotalAlpha, fmt.Sprintf("%.2f%%", stats.PersentaseAlpha)},
		}
		for _, dataRow := range statusData {
			if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), dataRow[0]); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("B%d", row), dataRow[1]); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("C%d", row), dataRow[2]); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			row++
		}
		row++ // Blank row after status table

		// Write average attendance
		if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), "Rata-rata Kehadiran"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		if subHeaderStyle != 0 {
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), subHeaderStyle); err != nil {
				logger.Warnw("Failed to set style", "error", err.Error())
			}
		}
		if err := setCellValue(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("%.2f%%", stats.RataRataKehadiran)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		row += 3 // Blank rows before detail section

		// Write detail section header
		if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), "DETAIL ABSENSI"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
			return
		}
		if subHeaderStyle != 0 {
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), subHeaderStyle); err != nil {
				logger.Warnw("Failed to set style", "error", err.Error())
			}
		}
		row++

		// Write detail table headers
		detailHeaders := []string{"No", "NIS", "Nama Siswa", "Kelas", "Jurusan", "Tanggal", "Status", "Deskripsi"}
		for colIndex, header := range detailHeaders {
			cellName, err := excelize.CoordinatesToCellName(colIndex+1, row)
			if err != nil {
				logger.Errorw("Failed to calculate cell name for detail header", "error", err.Error(), "column", colIndex+1)
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Gagal menyiapkan file Excel",
					"error":   "Internal error calculating cell coordinates",
				})
				return
			}
			if err := setCellValue(sheetName, cellName, header); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if tableHeaderStyle != 0 {
				if err := f.SetCellStyle(sheetName, cellName, cellName, tableHeaderStyle); err != nil {
					logger.Warnw("Failed to set style", "error", err.Error())
				}
			}
		}
		row++ // Move to next row for data

		// Write detail data rows
		for _, absensi := range absensiList {
			namaSiswa := ""
			kelas := ""
			jurusan := ""
			if absensi.Siswa != nil {
				namaSiswa = absensi.Siswa.NamaSiswa
				kelas = absensi.Siswa.Kelas
				jurusan = absensi.Siswa.Jurusan
			}

			if err := setCellValue(sheetName, fmt.Sprintf("A%d", row), row-1); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			} // No (row index starting from 1)
			if err := setCellValue(sheetName, fmt.Sprintf("B%d", row), absensi.NIS); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("C%d", row), namaSiswa); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("D%d", row), kelas); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("E%d", row), jurusan); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("F%d", row), absensi.Tanggal.Format("2006-01-02")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("G%d", row), absensi.Status); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			if err := setCellValue(sheetName, fmt.Sprintf("H%d", row), absensi.Deskripsi); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menulis file Excel"})
				return
			}
			row++
		}

		// Auto-adjust column widths
		colLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
		widths := []float64{5, 12, 20, 12, 15, 12, 12, 20}
		for i, colLetter := range colLetters {
			if err := f.SetColWidth(sheetName, colLetter, colLetter, widths[i]); err != nil {
				logger.Warnw("Failed to set column width", "error", err.Error())
			}
		}

		// Set response headers
		filename := fmt.Sprintf("laporan_absensi_%s.xlsx", time.Now().Format("20060102_150405"))
		c.Header("Content-Type", ExcelContentType)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		// Write to response
		if err := f.Write(c.Writer); err != nil {
			logger.Errorw("Failed to write Excel file to response",
				"error", err.Error(),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal mengirim file Excel",
				"error":   err.Error(),
			})
			return
		}

		logger.Infow("Laporan exported to Excel successfully",
			"total_records", len(absensiList),
			"stats", stats,
		)
	}
}
