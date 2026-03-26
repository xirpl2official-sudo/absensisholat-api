package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"absensholat-api/models"
	"absensholat-api/utils"
)

// CreateUserStaffRequest represents payload to create a staff user
type CreateUserStaffRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
	// Optional profile fields
	Name string `json:"name,omitempty"`
	NIP  string `json:"nip,omitempty"`
}

// UpdateUserStaffRequest represents payload to update a staff user
type UpdateUserStaffRequest struct {
	Username *string `json:"username,omitempty" binding:"omitempty,alphanum"`
	Password *string `json:"password,omitempty"`
	Role     *string `json:"role,omitempty"`
	Name     *string `json:"name,omitempty"`
	NIP      *string `json:"nip,omitempty"`
}

// SQLQueryRequest represents a safe read-only SQL query payload
type SQLQueryRequest struct {
	Query string `json:"query" binding:"required"`
}

// GenericDataResponse is a generic wrapper used when responses return arbitrary data in a `data` field
type GenericDataResponse struct {
	Data interface{} `json:"data"`
}

// AdminListUsersStaff godoc
// @Summary List staff users
// @Description Lists staff users (admin-only)
// @Tags admin
// @Produce json
// @Success 200 {array} models.UserStaff
// @Failure 500 {object} ErrorResponse
// @Router /admin/users_staff [get]
func AdminListUsersStaff(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var staff []models.UserStaff
		if err := db.Find(&staff).Error; err != nil {
			logger.Errorw("failed to list staff", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil daftar staff"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": staff})
	}
}

// AdminGetUserStaff godoc
// @Summary Get staff user
// @Description Get a staff user by id (admin-only)
// @Tags admin
// @Produce json
// @Param id path int true "ID staff"
// @Success 200 {object} models.UserStaff
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users_staff/{id} [get]
func AdminGetUserStaff(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var staff models.UserStaff
		if err := db.First(&staff, "id_staff = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Staff tidak ditemukan"})
				return
			}
			logger.Errorw("failed to get staff", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil staff"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": staff})
	}
}

// AdminCreateUserStaff godoc
// @Summary Create staff user
// @Description Create a staff user (admin-only)
// @Tags admin
// @Accept json
// @Produce json
// @Param request body CreateUserStaffRequest true "Staff payload"
// @Success 201 {object} models.UserStaff
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users_staff [post]
func AdminCreateUserStaff(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input CreateUserStaffRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Payload tidak valid", "error": err.Error()})
			return
		}

		// Validate role
		allowedRoles := map[string]bool{"admin": true, "guru": true, "wali_kelas": true}
		if !allowedRoles[input.Role] {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Role tidak valid"})
			return
		}

		// Check username uniqueness
		var existing models.UserStaff
		if err := db.First(&existing, "username = ?", input.Username).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"message": "Username sudah digunakan"})
			return
		} else if err != gorm.ErrRecordNotFound {
			logger.Errorw("db error checking username", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal memeriksa username"})
			return
		}

		// Validate password strength
		if err := utils.ValidatePassword(input.Password); err != nil {
			if pwdErr, ok := err.(*utils.PasswordValidationError); ok {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Password tidak memenuhi syarat", "errors": pwdErr.Errors})
				return
			}
		}

		hashed, err := utils.HashPassword(input.Password)
		if err != nil {
			logger.Errorw("failed to hash password", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal memproses password"})
			return
		}

		staff := models.UserStaff{
			Username: input.Username,
			Password: hashed,
			Role:     input.Role,
		}

		if err := db.Create(&staff).Error; err != nil {
			logger.Errorw("failed to create staff", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal membuat staff"})
			return
		}

		// Create profile record when provided
		if input.Role == "admin" && input.Name != "" {
			admin := models.Admin{IDStaff: staff.IDStaff, NamaAdmin: input.Name}
			db.Create(&admin) // ignore error for non-critical
		} else if input.Role == "guru" && input.Name != "" {
			guru := models.Guru{IDStaff: staff.IDStaff, NamaGuru: input.Name, NIP: input.NIP}
			db.Create(&guru)
		}

		c.JSON(http.StatusCreated, gin.H{"data": staff})
	}
}

// AdminUpdateUserStaff godoc
// @Summary Update staff user
// @Description Update a staff user (admin-only)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID staff"
// @Param request body UpdateUserStaffRequest true "Update payload"
// @Success 200 {object} models.UserStaff
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users_staff/{id} [put]
func AdminUpdateUserStaff(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var staff models.UserStaff
		if err := db.First(&staff, "id_staff = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Staff tidak ditemukan"})
				return
			}
			logger.Errorw("failed to fetch staff", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil staff"})
			return
		}

		var input UpdateUserStaffRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Payload tidak valid", "error": err.Error()})
			return
		}

		if input.Username != nil {
			staff.Username = *input.Username
		}
		if input.Password != nil {
			if err := utils.ValidatePassword(*input.Password); err != nil {
				if pwdErr, ok := err.(*utils.PasswordValidationError); ok {
					c.JSON(http.StatusBadRequest, gin.H{"message": "Password tidak memenuhi syarat", "errors": pwdErr.Errors})
					return
				}
			}
			hashed, err := utils.HashPassword(*input.Password)
			if err != nil {
				logger.Errorw("failed to hash password", "error", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal memproses password"})
				return
			}
			staff.Password = hashed
		}
		if input.Role != nil {
			staff.Role = *input.Role
		}

		if err := db.Save(&staff).Error; err != nil {
			logger.Errorw("failed to update staff", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal memperbarui staff"})
			return
		}

		// Update profile if provided
		if input.Name != nil || input.NIP != nil {
			if staff.Role == "admin" {
				var admin models.Admin
				if err := db.First(&admin, "id_staff = ?", staff.IDStaff).Error; err == nil {
					if input.Name != nil {
						admin.NamaAdmin = *input.Name
					}
					db.Save(&admin)
				} else if input.Name != nil {
					admin = models.Admin{IDStaff: staff.IDStaff, NamaAdmin: *input.Name}
					db.Create(&admin)
				}
			} else if staff.Role == "guru" {
				var guru models.Guru
				if err := db.First(&guru, "id_staff = ?", staff.IDStaff).Error; err == nil {
					if input.Name != nil {
						guru.NamaGuru = *input.Name
					}
					if input.NIP != nil {
						guru.NIP = *input.NIP
					}
					db.Save(&guru)
				} else {
					newGuru := models.Guru{IDStaff: staff.IDStaff}
					if input.Name != nil {
						newGuru.NamaGuru = *input.Name
					}
					if input.NIP != nil {
						newGuru.NIP = *input.NIP
					}
					db.Create(&newGuru)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": staff})
	}
}

// AdminDeleteUserStaff godoc
// @Summary Delete staff user
// @Description Delete a staff user and related profile records (admin-only)
// @Tags admin
// @Param id path int true "ID staff"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users_staff/{id} [delete]
func AdminDeleteUserStaff(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var staff models.UserStaff
		if err := db.First(&staff, "id_staff = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Staff tidak ditemukan"})
				return
			}
			logger.Errorw("failed to fetch staff", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil staff"})
			return
		}

		// Delete related profiles (ignore errors)
		db.Where("id_staff = ?", staff.IDStaff).Delete(&models.Admin{})
		db.Where("id_staff = ?", staff.IDStaff).Delete(&models.Guru{})

		if err := db.Delete(&staff).Error; err != nil {
			logger.Errorw("failed to delete staff", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menghapus staff"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// AdminRunSelectQuery godoc
// @Summary Run a read-only SELECT query
// @Description Execute a single SELECT SQL query and return results (admin-only). Only single SELECT queries without semicolons are allowed.
// @Tags admin
// @Accept json
// @Produce json
// @Param request body SQLQueryRequest true "Query payload"
// @Success 200 {object} GenericDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/query [post]
func AdminRunSelectQuery(db *gorm.DB, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SQLQueryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Payload tidak valid", "error": err.Error()})
			return
		}

		q := strings.TrimSpace(req.Query)
		low := strings.ToLower(q)
		// Simple safety checks
		if !strings.HasPrefix(low, "select") {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Hanya query SELECT yang diizinkan"})
			return
		}
		if strings.Contains(q, ";") {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Beberapa pernyataan tidak diizinkan"})
			return
		}

		rows, err := db.Raw(q).Rows()
		if err != nil {
			logger.Errorw("query error", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menjalankan query"})
			return
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			logger.Errorw("failed to get columns", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal membaca hasil"})
			return
		}

		results := make([]map[string]interface{}, 0)
		for rows.Next() {
			colsPtrs := make([]interface{}, len(cols))
			colsVals := make([]sql.NullString, len(cols))
			for i := range colsPtrs {
				colsPtrs[i] = &colsVals[i]
			}
			if err := rows.Scan(colsPtrs...); err != nil {
				logger.Errorw("row scan error", "error", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal membaca baris hasil"})
				return
			}
			rowMap := make(map[string]interface{})
			for i, col := range cols {
				if colsVals[i].Valid {
					rowMap[col] = colsVals[i].String
				} else {
					rowMap[col] = nil
				}
			}
			results = append(results, rowMap)
		}

		c.JSON(http.StatusOK, gin.H{"data": results})
	}
}
