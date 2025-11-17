package postgres

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	user := &entities.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(1, now, now)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, user.PasswordHash).
		WillReturnRows(rows)

	err = repo.Create(user)
	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	expectedUser := &entities.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PasswordHash, now, now)

	mock.ExpectQuery(`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id`).
		WithArgs(1).
		WillReturnRows(rows)

	user, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	expectedUser := &entities.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PasswordHash, now, now)

	mock.ExpectQuery(`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	user, err := repo.GetByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	user := &entities.User{
		ID:           1,
		Email:        "updated@example.com",
		PasswordHash: "newhashedpassword",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"updated_at"}).
		AddRow(now)

	mock.ExpectQuery(`UPDATE users SET email`).
		WithArgs(user.Email, user.PasswordHash, user.ID).
		WillReturnRows(rows)

	err = repo.Update(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
