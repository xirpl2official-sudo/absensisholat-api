package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"absensholat-api/models"
)

type NotificationItem struct {
	NIS         string `json:"nis"`
	NamaSiswa   string `json:"nama_siswa"`
	Kelas       string `json:"kelas"`
	Jurusan     string `json:"jurusan"`
	JenisSholat string `json:"jenis_sholat"`
	WaktuMulai  string `json:"waktu_mulai"`
	IDJadwal    int    `json:"id_jadwal"`
}

type NotificationResponse struct {
	Message string             `json:"message"`
	Data    []NotificationItem `json:"data"`
	Count   int                `json:"count"`
}

// GetPendingNotifications returns students who haven't marked attendance for active/ended prayers today
// @Summary Get pending attendance notifications
// @Description Returns students who need to mark attendance for prayers that have started or ended today
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} NotificationResponse
// @Failure 500 {object} ErrorResponse
// @Router /notifications [get]
func GetPendingNotifications(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current time in WIB
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			loc = time.FixedZone("WIB", 7*3600)
			logger.Warnw("failed to load Asia/Jakarta timezone, using fixed fallback", "error", err.Error())
		}
		now := time.Now().In(loc)
		currentDay := getDayName(now.Weekday())
		currentTime := now.Format("15:04:05")
		today := now.Format("2006-01-02")

		// Find prayers that have started (waktu_mulai <= now)
		var activePrayers []models.JadwalSholat
		if filterErr := db.Where("hari = ? AND waktu_mulai <= ?",
			currentDay, currentTime).
			Find(&activePrayers).Error; filterErr != nil {
			logger.Errorw("Failed to fetch active prayers", "error", filterErr.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal mengambil data jadwal",
			})
			return
		}

		if len(activePrayers) == 0 {
			c.JSON(http.StatusOK, NotificationResponse{
				Message: "Tidak ada jadwal sholat aktif",
				Data:    []NotificationItem{},
				Count:   0,
			})
			return
		}

		// Get prayer IDs
		prayerIDs := make([]int, len(activePrayers))
		prayerMap := make(map[int]models.JadwalSholat)
		for i, p := range activePrayers {
			prayerIDs[i] = p.IDJadwal
			prayerMap[p.IDJadwal] = p
		}

		// Get all students
		var students []models.Siswa
		if findErr := db.Find(&students).Error; findErr != nil {
			logger.Errorw("Failed to fetch students", "error", findErr.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal mengambil data siswa",
			})
			return
		}

		// Get existing attendance records for today
		tanggal, err := time.Parse("2006-01-02", today)
		if err != nil {
			logger.Errorw("Failed to parse today's date", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Kesalahan internal (tanggal)",
			})
			return
		}
		var existingRecords []models.Absensi
		db.Where("tanggal = ? AND id_jadwal IN (?)", tanggal, prayerIDs).Find(&existingRecords)

		existingMap := make(map[string]bool)
		for _, r := range existingRecords {
			key := r.NIS + "-" + string(rune(r.IDJadwal))
			existingMap[key] = true
		}

		// Use a more reliable key format
		existingSet := make(map[string]bool)
		for _, r := range existingRecords {
			existingSet[r.NIS+"|"+time.Time(r.Tanggal).Format("2006-01-02")+"|"+string(rune(r.IDJadwal))] = true
		}

		// Simplified: just use NIS-IDJadwal
		attended := make(map[string]bool)
		for _, r := range existingRecords {
			attended[r.NIS+"-"+itoa(r.IDJadwal)] = true
		}

		var notifications []NotificationItem
		for _, prayer := range activePrayers {
			for _, student := range students {
				key := student.NIS + "-" + itoa(prayer.IDJadwal)
				if !attended[key] {
					// Filter Dhuha by jurusan
					if prayer.JenisSholat == "Dhuha" && prayer.Jurusan != "" && prayer.Jurusan != "Semua Jurusan" {
						if student.Jurusan != prayer.Jurusan {
							continue
						}
					}
					// Gender-based Friday filtering
					if now.Weekday() == time.Friday {
						if student.JK == "L" && prayer.JenisSholat == "Dzuhur" {
							continue
						}
						if student.JK == "P" && prayer.JenisSholat == "Jumat" {
							continue
						}
					}

					notifications = append(notifications, NotificationItem{
						NIS:         student.NIS,
						NamaSiswa:   student.NamaSiswa,
						Kelas:       student.Kelas,
						Jurusan:     student.Jurusan,
						JenisSholat: prayer.JenisSholat,
						WaktuMulai:  prayer.WaktuMulai,
						IDJadwal:    prayer.IDJadwal,
					})
				}
			}
		}

		c.JSON(http.StatusOK, NotificationResponse{
			Message: "Notifikasi kehadiran berhasil diambil",
			Data:    notifications,
			Count:   len(notifications),
		})
	}
}

// itoa converts int to string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var s string
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
