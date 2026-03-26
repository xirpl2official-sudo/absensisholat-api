package handlers

import (
	"absensholat-api/scheduler"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TriggerAutoAbsen triggers the automatic alpha attendance recording process manually.
// @Summary Trigger auto absen alpha manual
// @Description Trigger the process to mark students as alpha for ended prayers. Usually called by Vercel Crons.
// @Tags auto-absen
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auto-absen/trigger [post]
func TriggerAutoAbsen(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Note: Auth middleware (admin only) is applied in routes

		err := scheduler.AutoAbsenAlpha(db, logger)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal menjalankan auto-absen",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Auto-absen alpha berhasil dijalankan",
		})
	}
}
