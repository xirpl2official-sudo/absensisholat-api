package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"absensholat-api/models"
	"absensholat-api/utils"
)

// ChangeEmailRequest represents the change email request body
type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

// VerifyEmailOTPRequest represents the email OTP verification request body
type VerifyEmailOTPRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	OTP      string `json:"otp" binding:"required"`
}

// RequestChangeEmail godoc
// @Summary Request Email Change
// @Description Request OTP untuk mengubah email. OTP akan dikirim ke email baru untuk verifikasi
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangeEmailRequest true "Email baru untuk akun"
// @Success 200 {object} map[string]interface{} "OTP berhasil dikirim ke email baru"
// @Failure 400 {object} ErrorResponse "Permintaan tidak valid - Email kosong, format salah, atau bukan email Google"
// @Failure 401 {object} ErrorResponse "Tidak terotentikasi - Token tidak valid"
// @Failure 409 {object} ErrorResponse "Email sudah digunakan oleh akun lain"
// @Failure 500 {object} ErrorResponse "Gagal mengirim email atau error server"
// @Router /auth/change-email [post]
func RequestChangeEmail(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get NIS from auth middleware
		nis, exists := c.Get("nis")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "User not authenticated",
			})
			return
		}

		var input ChangeEmailRequest

		if err := c.ShouldBindJSON(&input); err != nil {
			logger.Warnw("Invalid JSON format for change email",
				"error", err.Error(),
				"ip", c.ClientIP(),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Format JSON tidak valid",
				"error":   err.Error(),
			})
			return
		}

		logger.Infow("Change email request received",
			"nis", nis,
			"new_email", input.NewEmail,
		)

		// Validate email format
		if !isValidEmail(input.NewEmail) {
			logger.Warnw("Invalid email format",
				"email", input.NewEmail,
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Format email tidak valid",
			})
			return
		}

		// Validate Google account
		if !isGoogleAccount(input.NewEmail) {
			logger.Warnw("Non-Google email provided",
				"email", input.NewEmail,
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Email harus menggunakan domain Google (contoh: @gmail.com)",
			})
			return
		}

		// Get current account
		var account models.AkunLoginSiswa
		if err := db.Preload("Siswa").First(&account, "nis = ?", nis).Error; err != nil {
			logger.Errorw("Database error during change email",
				"error", err.Error(),
				"nis", nis,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal memproses permintaan",
			})
			return
		}

		// Check if new email is same as current email
		if account.Email == input.NewEmail {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Email baru sama dengan email saat ini",
			})
			return
		}

		// Check if new email is already used by another account
		var existingAccount models.AkunLoginSiswa
		if err := db.First(&existingAccount, "email = ?", input.NewEmail).Error; err == nil {
			logger.Warnw("Email already in use",
				"email", input.NewEmail,
			)
			c.JSON(http.StatusConflict, gin.H{
				"message": "Email sudah digunakan oleh akun lain",
			})
			return
		} else if err != gorm.ErrRecordNotFound {
			logger.Errorw("Database error checking email",
				"error", err.Error(),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal memverifikasi email",
			})
			return
		}

		// Generate OTP
		otpCode, err := utils.GenerateOTP()
		if err != nil {
			logger.Errorw("Failed to generate OTP",
				"error", err.Error(),
				"nis", nis,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal membuat kode OTP",
			})
			return
		}

		// Store OTP in Firebase with 10 minutes expiration
		// Use a special key for email change: "email_change:{nis}:{new_email}"
		otpStore := utils.GetOTPStore()
		if otpStore == nil {
			logger.Errorw("OTP store not initialized",
				"nis", nis,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Layanan OTP tidak tersedia. Pastikan Firebase sudah dikonfigurasi",
			})
			return
		}

		// Store with special identifier for email change
		nisStr, ok := nis.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Format NIS tidak valid"})
			return
		}
		emailChangeKey := nisStr + "_email_change"
		if err := otpStore.SaveOTP(emailChangeKey, input.NewEmail, otpCode, 10*time.Minute); err != nil {
			logger.Errorw("Failed to save OTP to Firebase",
				"error", err.Error(),
				"nis", nis,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal menyimpan kode OTP",
			})
			return
		}

		// Get student name for email
		namaSiswa := "Siswa"
		if account.Siswa != nil {
			namaSiswa = account.Siswa.NamaSiswa
		}

		// Send OTP email to new email address
		if err := utils.SendOTPEmail(input.NewEmail, otpCode, namaSiswa); err != nil {
			logger.Errorw("Failed to send OTP email",
				"error", err.Error(),
				"nis", nis,
				"email", input.NewEmail,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal mengirim email OTP",
				"error":   err.Error(),
			})
			return
		}

		logger.Infow("Email change OTP generated and sent successfully",
			"nis", nis,
			"new_email", input.NewEmail,
		)

		c.JSON(http.StatusOK, gin.H{
			"message":    "Kode OTP telah dikirim ke email baru Anda",
			"email":      maskEmail(input.NewEmail),
			"expires_in": "10 menit",
		})
	}
}

// VerifyAndChangeEmail godoc
// @Summary Verify OTP and Change Email
// @Description Verifikasi kode OTP dan mengubah email akun
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body VerifyEmailOTPRequest true "Email baru dan kode OTP"
// @Success 200 {object} map[string]interface{} "Email berhasil diubah"
// @Failure 400 {object} ErrorResponse "OTP tidak valid atau sudah kadaluarsa"
// @Failure 401 {object} ErrorResponse "Tidak terotentikasi - Token tidak valid"
// @Failure 500 {object} ErrorResponse "Gagal mengubah email"
// @Router /auth/verify-change-email [post]
func VerifyAndChangeEmail(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get NIS from auth middleware
		nis, exists := c.Get("nis")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "User not authenticated",
			})
			return
		}

		var input VerifyEmailOTPRequest

		if err := c.ShouldBindJSON(&input); err != nil {
			logger.Warnw("Invalid JSON format for verify email change",
				"error", err.Error(),
				"ip", c.ClientIP(),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Format JSON tidak valid",
				"error":   err.Error(),
			})
			return
		}

		logger.Infow("Verify email change request received",
			"nis", nis,
			"new_email", input.NewEmail,
		)

		// Validate email format
		if !isValidEmail(input.NewEmail) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Format email tidak valid",
			})
			return
		}

		// Validate Google account
		if !isGoogleAccount(input.NewEmail) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Email harus menggunakan domain Google (contoh: @gmail.com)",
			})
			return
		}

		otpStore := utils.GetOTPStore()
		if otpStore == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Layanan OTP tidak tersedia",
			})
			return
		}

		// Verify OTP with special email change key
		nisStr, ok := nis.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Format NIS tidak valid"})
			return
		}
		emailChangeKey := nisStr + "_email_change"
		valid, err := otpStore.VerifyOTP(emailChangeKey, input.OTP)

		if err != nil {
			logger.Warnw("OTP verification failed",
				"nis", nis,
				"error", err.Error(),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Kode OTP tidak valid",
			})
			return
		}

		// Check if email is still available (in case another user took it)
		var existingAccount models.AkunLoginSiswa
		if err := db.First(&existingAccount, "email = ?", input.NewEmail).Error; err == nil {
			logger.Warnw("Email already in use during verification",
				"email", input.NewEmail,
			)
			c.JSON(http.StatusConflict, gin.H{
				"message": "Email sudah digunakan oleh akun lain",
			})
			return
		} else if err != gorm.ErrRecordNotFound {
			logger.Errorw("Database error checking email",
				"error", err.Error(),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal memverifikasi email",
			})
			return
		}

		// Update email in database
		result := db.Model(&models.AkunLoginSiswa{}).
			Where("nis = ?", nis).
			Update("email", input.NewEmail)

		if result.Error != nil {
			logger.Errorw("Failed to update email",
				"error", result.Error.Error(),
				"nis", nis,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal menyimpan email baru",
			})
			return
		}

		if result.RowsAffected == 0 {
			logger.Warnw("No account found to update email",
				"nis", nis,
			)
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Akun tidak ditemukan",
			})
			return
		}

		// Clear OTP after successful email change
		if err := otpStore.ClearOTP(emailChangeKey); err != nil {
			logger.Warnw("Failed to clear OTP from Firebase",
				"error", err.Error(),
				"nis", nis,
			)
			// Don't fail the request, email was already changed
		}

		logger.Infow("Email changed successfully",
			"nis", nis,
			"new_email", input.NewEmail,
		)

		c.JSON(http.StatusOK, gin.H{
			"message":   "Email berhasil diubah",
			"new_email": input.NewEmail,
		})
	}
}
