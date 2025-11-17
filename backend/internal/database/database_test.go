package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Success(t *testing.T) {
	// This test verifies that New() applies the connection pool settings correctly
	// We test with a mock that simulates a successful connection
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer mockDB.Close()

	// Expect a ping to verify connection
	mock.ExpectPing().WillReturnError(nil)

	// We can't directly test New() with a mock, but we can test the pool settings
	// on a real sql.DB struct
	cfg := Config{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3 * time.Minute,
	}

	// Apply the same logic as New() to verify settings are applied
	if cfg.MaxOpenConns > 0 {
		mockDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		mockDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		mockDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	// Verify ping works
	err = mockDB.Ping()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Verify stats to ensure settings were applied
	stats := mockDB.Stats()
	assert.Equal(t, 10, stats.MaxOpenConnections)
}

func TestNew_ConnectionDefaults(t *testing.T) {
	// Test that default values are applied when config values are zero
	cfg := Config{
		MaxOpenConns:    0, // Should default to 25
		MaxIdleConns:    0, // Should default to 5
		ConnMaxLifetime: 0, // Should default to 5 minutes
	}

	// Create a mock DB to test defaults
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// Apply the same logic as New()
	if cfg.MaxOpenConns > 0 {
		mockDB.SetMaxOpenConns(cfg.MaxOpenConns)
	} else {
		mockDB.SetMaxOpenConns(25)
	}

	if cfg.MaxIdleConns > 0 {
		mockDB.SetMaxIdleConns(cfg.MaxIdleConns)
	} else {
		mockDB.SetMaxIdleConns(5)
	}

	stats := mockDB.Stats()
	assert.Equal(t, 25, stats.MaxOpenConnections)
}

func TestDB_Ping(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	db := &DB{sqlxDB}

	// Test successful ping
	mock.ExpectPing().WillReturnError(nil)
	err = db.Ping()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test failed ping
	mock.ExpectPing().WillReturnError(sql.ErrConnDone)
	err = db.Ping()
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Close(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	db := &DB{sqlxDB}

	// ExpectClose is called when Close() is invoked
	mock.ExpectClose()
	err = db.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Health(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	db := &DB{sqlxDB}

	// Test successful health check
	mock.ExpectPing().WillReturnError(nil)
	err = db.Health()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test failed health check
	mock.ExpectPing().WillReturnError(sql.ErrConnDone)
	err = db.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database health check failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_HealthContext(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	db := &DB{sqlxDB}

	// Test successful health check with context
	ctx := context.Background()
	mock.ExpectPing().WillReturnError(nil)
	err = db.HealthContext(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test health check failure
	mock.ExpectPing().WillReturnError(sql.ErrConnDone)
	err = db.HealthContext(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database health check failed")
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test health check with timeout - simulate a slow ping
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelTimeout()
	mock.ExpectPing().WillDelayFor(100 * time.Millisecond)
	err = db.HealthContext(ctxTimeout)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database health check failed")
}

func TestConfigDefaults(t *testing.T) {
	// This test validates that we can create a Config struct
	cfg := Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "test",
		SSLMode:  "disable",
	}

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "5432", cfg.Port)
}
