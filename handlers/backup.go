package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BackupLog represents a backup tracking record
type BackupLog struct {
	ID              int        `gorm:"primaryKey" json:"id"`
	StartDate       time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate         time.Time  `gorm:"type:date;not null" json:"end_date"`
	FileFormat      string     `gorm:"default:xlsx" json:"file_format"`
	ExportedAt      time.Time  `gorm:"autoCreateTime" json:"exported_at"`
	AutoDeleteAfter *time.Time `json:"auto_delete_after"`
	DeletedAt       *time.Time `json:"deleted_at"`
	CreatedBy       *int       `json:"created_by"`
}

func (BackupLog) TableName() string {
	return "backup_log"
}

type BackupStatusResponse struct {
	Message       string      `json:"message"`
	HasPending    bool        `json:"has_pending"`
	PendingRanges []DateRange `json:"pending_ranges"`
	RecentBackups []BackupLog `json:"recent_backups"`
}

type DateRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Count     int64  `json:"count"`
}

// GetBackupStatus returns dates with un-exported attendance data
// @Summary Get backup status
// @Description Returns date ranges with attendance data that hasn't been backed up
// @Tags backup
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} BackupStatusResponse
// @Failure 500 {object} ErrorResponse
// @Router /backup/status [get]
func GetBackupStatus(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Find dates with absensi data that haven't been backed up
		type DateCount struct {
			Tanggal time.Time
			Count   int64
		}

		var dateCounts []DateCount
		if err := db.Raw(`
			SELECT DATE(tanggal) as tanggal, COUNT(*) as count
			FROM absensi
			WHERE DATE(tanggal) NOT IN (
				SELECT DISTINCT d::date
				FROM backup_log bl,
				     generate_series(bl.start_date, bl.end_date, '1 day'::interval) d
				WHERE bl.deleted_at IS NULL
			)
			GROUP BY DATE(tanggal)
			ORDER BY DATE(tanggal) DESC
			LIMIT 30
		`).Scan(&dateCounts).Error; err != nil {
			// If backup_log table doesn't exist yet, fallback
			logger.Warnw("Failed to check backup status, may need migration", "error", err.Error())

			// Fallback: show all dates with data
			if err2 := db.Raw(`
				SELECT DATE(tanggal) as tanggal, COUNT(*) as count
				FROM absensi
				GROUP BY DATE(tanggal)
				ORDER BY DATE(tanggal) DESC
				LIMIT 30
			`).Scan(&dateCounts).Error; err2 != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil status backup"})
				return
			}
		}

		pendingRanges := make([]DateRange, 0)
		for _, dc := range dateCounts {
			dateStr := dc.Tanggal.Format("2006-01-02")
			pendingRanges = append(pendingRanges, DateRange{
				StartDate: dateStr,
				EndDate:   dateStr,
				Count:     dc.Count,
			})
		}

		// Get recent backups
		var recentBackups []BackupLog
		db.Where("deleted_at IS NULL").Order("exported_at DESC").Limit(10).Find(&recentBackups)

		c.JSON(http.StatusOK, BackupStatusResponse{
			Message:       "Status backup berhasil diambil",
			HasPending:    len(pendingRanges) > 0,
			PendingRanges: pendingRanges,
			RecentBackups: recentBackups,
		})
	}
}

type ConfirmBackupRequest struct {
	StartDate  string `json:"start_date" binding:"required"`
	EndDate    string `json:"end_date" binding:"required"`
	FileFormat string `json:"file_format"`
}

// ConfirmBackup marks a date range as backed up
// @Summary Confirm backup completion
// @Description Marks a date range as successfully backed up
// @Tags backup
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ConfirmBackupRequest true "Backup confirmation"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /backup/confirm [post]
func ConfirmBackup(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ConfirmBackupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Data tidak valid", "error": err.Error()})
			return
		}

		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Format start_date tidak valid"})
			return
		}
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Format end_date tidak valid"})
			return
		}

		fileFormat := req.FileFormat
		if fileFormat == "" {
			fileFormat = "xlsx"
		}

		// Auto-delete 24 hours after backup
		autoDeleteAfter := time.Now().Add(24 * time.Hour)

		// Get staff ID from context
		var createdBy *int
		if idStaff, exists := c.Get("id_staff"); exists {
			if id, ok := idStaff.(int); ok {
				createdBy = &id
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Format staff ID tidak valid"})
				return
			}
		}

		log := BackupLog{
			StartDate:       startDate,
			EndDate:         endDate,
			FileFormat:      fileFormat,
			AutoDeleteAfter: &autoDeleteAfter,
			CreatedBy:       createdBy,
		}

		if err := db.Create(&log).Error; err != nil {
			logger.Errorw("Failed to create backup log", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyimpan log backup"})
			return
		}

		logger.Infow("Backup confirmed",
			"start_date", req.StartDate,
			"end_date", req.EndDate,
			"auto_delete_after", autoDeleteAfter,
		)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Backup berhasil dikonfirmasi",
			"data":    log,
		})
	}
}

// CleanupBackedUpData deletes attendance data that was backed up more than 24h ago
// @Summary Cleanup backed up data
// @Description Deletes attendance records for date ranges where backup was confirmed > 24h ago
// @Tags backup
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /backup/cleanup [delete]
func CleanupBackedUpData(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()

		// Find backup logs where auto_delete_after has passed and data hasn't been deleted
		var logs []BackupLog
		if err := db.Where("auto_delete_after <= ? AND deleted_at IS NULL", now).Find(&logs).Error; err != nil {
			logger.Errorw("Failed to fetch cleanup candidates", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil data cleanup"})
			return
		}

		totalDeleted := int64(0)
		for _, log := range logs {
			// Delete absensi records in this date range
			result := db.Where("DATE(tanggal) >= ? AND DATE(tanggal) <= ?",
				log.StartDate, log.EndDate).Delete(&struct {
				IDAbsen int `gorm:"primaryKey;column:id_absen"`
			}{})

			if result.Error != nil {
				logger.Errorw("Failed to delete absensi records",
					"error", result.Error.Error(),
					"start_date", log.StartDate,
					"end_date", log.EndDate,
				)
				continue
			}

			totalDeleted += result.RowsAffected

			// Mark backup log as deleted
			deletedAt := now
			db.Model(&log).Updates(map[string]interface{}{"deleted_at": &deletedAt})
		}

		logger.Infow("Cleanup completed",
			"logs_processed", len(logs),
			"records_deleted", totalDeleted,
		)

		c.JSON(http.StatusOK, gin.H{
			"message":         "Cleanup berhasil dilakukan",
			"logs_processed":  len(logs),
			"records_deleted": totalDeleted,
		})
	}
}

// StartBackupCleanupScheduler runs cleanup every hour in the background
func StartBackupCleanupScheduler(db *gorm.DB, logger *zap.SugaredLogger) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			var logs []BackupLog
			if err := db.Where("auto_delete_after <= ? AND deleted_at IS NULL", now).Find(&logs).Error; err != nil {
				logger.Errorw("Backup cleanup scheduler failed", "error", err.Error())
				continue
			}

			for _, log := range logs {
				result := db.Exec(
					"DELETE FROM absensi WHERE DATE(tanggal) >= ? AND DATE(tanggal) <= ?",
					log.StartDate, log.EndDate,
				)

				if result.Error != nil {
					logger.Errorw("Auto-cleanup failed",
						"error", result.Error.Error(),
						"log_id", log.ID,
					)
					continue
				}

				deletedAt := now
				db.Model(&log).Updates(map[string]interface{}{"deleted_at": &deletedAt})

				logger.Infow("Auto-cleanup completed",
					"log_id", log.ID,
					"records_deleted", result.RowsAffected,
					"date_range", fmt.Sprintf("%s to %s", log.StartDate.Format("2006-01-02"), log.EndDate.Format("2006-01-02")),
				)
			}
		}
	}()
}
