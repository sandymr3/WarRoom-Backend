package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// Helper to open DB from connector since sql.OpenDB is available in Go 1.10+
func openDBWithConnector(connector driver.Connector) *sql.DB {
	return sql.OpenDB(connector)
}

// WithRetry executes the given function with a fixed number of retries for connection issues.
func WithRetry(fn func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		// Only retry on connection-related errors
		msg := err.Error()
		if isConnectionError(msg) {
			log.Printf("DB connection error, retrying (%d/3): %v", i+1, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return err
	}
	return fmt.Errorf("failed after 3 retries: %w", err)
}

func isConnectionError(msg string) bool {
	return (msg == "bad connection" ||
		msg == "invalid connection" ||
		fmt.Sprint(msg) == "connection forcibly closed by remote host" ||
		fmt.Sprint(msg) == "unexpected EOF")
}

// QueryTimer wraps a GORM query with timing logic.
func QueryTimer(label string, query *gorm.DB) *gorm.DB {
	start := time.Now()
	res := query
	duration := time.Since(start)
	if duration > 200*time.Millisecond {
		log.Printf("[SLOW QUERY] %s took %v", label, duration)
	}
	return res
}

// TransactionWithRetry executes a function within a transaction with retries.
func TransactionWithRetry(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return WithRetry(func() error {
		return DB.WithContext(ctx).Transaction(fn)
	})
}
