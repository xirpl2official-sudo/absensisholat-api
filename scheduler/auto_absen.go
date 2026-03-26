package scheduler

import (
	"absensholat-api/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AutoAbsenAlpha is a wrapper function to trigger the missed prayer recording logic.
// This is designed to be called by an API handler which is triggered by an external cron (like Vercel Crons).
func AutoAbsenAlpha(db *gorm.DB, logger *zap.SugaredLogger) error {
	logger.Info("[Scheduler] Starting automatic alpha attendance check...")

	// Delegate to the robust existing logic in utils
	err := utils.RecordMissedPrayers(db, logger)
	if err != nil {
		logger.Errorw("[Scheduler] Failed to record missed prayers", "error", err.Error())
		return err
	}

	logger.Info("[Scheduler] Automatic alpha attendance check completed successfully")
	return nil
}
