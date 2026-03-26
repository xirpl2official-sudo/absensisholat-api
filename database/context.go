package database

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// DB wraps gorm.DB to add context and metrics support
type DB struct {
	*gorm.DB
}

// Wrap creates a new DB wrapper
func Wrap(db *gorm.DB) *DB {
	return &DB{DB: db}
}

// WithContext adds request context to the database query
// This enables request cancellation and timeout enforcement
func (d *DB) WithContext(ctx context.Context) *gorm.DB {
	return d.DB.WithContext(ctx)
}

// Transaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
func (d *DB) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return d.DB.WithContext(ctx).Transaction(fn)
}

// QueryWithTimeout executes a query with a timeout
func QueryWithTimeout(db *gorm.DB, ctx context.Context, timeout time.Duration) *gorm.DB {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return db.WithContext(ctxWithTimeout)
}

// DefaultQueryTimeout is the default timeout for database queries
const DefaultQueryTimeout = 30 * time.Second

// WithDefaultTimeout adds the default timeout to the database query
func WithDefaultTimeout(db *gorm.DB, ctx context.Context) *gorm.DB {
	return QueryWithTimeout(db, ctx, DefaultQueryTimeout)
}
