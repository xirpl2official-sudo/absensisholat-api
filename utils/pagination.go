package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Pagination holds pagination parameters and results
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginationConfig holds configuration for pagination
type PaginationConfig struct {
	DefaultPage  int
	DefaultLimit int
	MaxLimit     int
}

// DefaultPaginationConfig returns default pagination configuration
func DefaultPaginationConfig() PaginationConfig {
	return PaginationConfig{
		DefaultPage:  1,
		DefaultLimit: 10,
		MaxLimit:     100,
	}
}

// NewPagination creates a new Pagination from gin context
func NewPagination(c *gin.Context) *Pagination {
	return NewPaginationWithConfig(c, DefaultPaginationConfig())
}

// NewPaginationWithConfig creates a new Pagination with custom config
func NewPaginationWithConfig(c *gin.Context, config PaginationConfig) *Pagination {
	page := config.DefaultPage
	limit := config.DefaultLimit

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > config.MaxLimit {
				limit = config.MaxLimit
			}
		}
	}

	return &Pagination{
		Page:  page,
		Limit: limit,
	}
}

// Offset returns the offset for database queries
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

// Apply applies pagination to a GORM query and counts total
func (p *Pagination) Apply(db *gorm.DB, model interface{}) *gorm.DB {
	var total int64
	db.Model(model).Count(&total)
	p.SetTotal(total)

	return db.Offset(p.Offset()).Limit(p.Limit)
}

// SetTotal sets the total count and calculates pagination metadata
func (p *Pagination) SetTotal(total int64) {
	p.Total = total
	p.TotalPages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.HasNext = p.Page < p.TotalPages
	p.HasPrev = p.Page > 1
}

// PaginatedResponse wraps data with pagination metadata
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// NewPaginatedResponse creates a paginated response
func NewPaginatedResponse(data interface{}, pagination *Pagination) PaginatedResponse {
	return PaginatedResponse{
		Data:       data,
		Pagination: *pagination,
	}
}

// SortOrder represents sort direction
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// SortConfig holds sorting configuration
type SortConfig struct {
	Field string
	Order SortOrder
}

// GetSortConfig extracts sort configuration from query parameters
func GetSortConfig(c *gin.Context, allowedFields []string, defaultField string, defaultOrder SortOrder) SortConfig {
	field := c.Query("sort_by")
	order := c.Query("order")

	// Validate field is allowed
	validField := false
	for _, f := range allowedFields {
		if f == field {
			validField = true
			break
		}
	}
	if !validField {
		field = defaultField
	}

	// Validate order
	sortOrder := defaultOrder
	if order == "asc" {
		sortOrder = SortAsc
	} else if order == "desc" {
		sortOrder = SortDesc
	}

	return SortConfig{
		Field: field,
		Order: sortOrder,
	}
}

// ApplySort applies sorting to a GORM query
func (s SortConfig) ApplySort(db *gorm.DB) *gorm.DB {
	if s.Field != "" {
		return db.Order(s.Field + " " + string(s.Order))
	}
	return db
}
