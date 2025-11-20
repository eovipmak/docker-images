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

func TestAlertChannelRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertChannelRepository(sqlxDB)

	channel := &entities.AlertChannel{
		TenantID: 1,
		Type:     "webhook",
		Name:     "Test Webhook",
		Config: entities.ChannelConfig{
			"url": "https://example.com/webhook",
		},
		Enabled: true,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow("channel-uuid", now, now)

	mock.ExpectQuery(`INSERT INTO alert_channels`).
		WithArgs(channel.TenantID, channel.Type, channel.Name, sqlmock.AnyArg(), channel.Enabled).
		WillReturnRows(rows)

	err = repo.Create(channel)
	assert.NoError(t, err)
	assert.Equal(t, "channel-uuid", channel.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertChannelRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertChannelRepository(sqlxDB)

	expectedChannel := &entities.AlertChannel{
		ID:       "channel-uuid",
		TenantID: 1,
		Type:     "discord",
		Name:     "Test Discord",
		Config: entities.ChannelConfig{
			"webhook_url": "https://discord.com/api/webhooks/xxx",
		},
		Enabled: true,
	}

	now := time.Now()
	configJSON := []byte(`{"webhook_url":"https://discord.com/api/webhooks/xxx"}`)
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "type", "name", "config", "enabled", "created_at", "updated_at"}).
		AddRow(expectedChannel.ID, expectedChannel.TenantID, expectedChannel.Type, expectedChannel.Name, configJSON, expectedChannel.Enabled, now, now)

	mock.ExpectQuery(`SELECT id, tenant_id, type, name, config, enabled, created_at, updated_at FROM alert_channels WHERE id`).
		WithArgs("channel-uuid").
		WillReturnRows(rows)

	channel, err := repo.GetByID("channel-uuid")
	assert.NoError(t, err)
	assert.Equal(t, expectedChannel.Name, channel.Name)
	assert.Equal(t, expectedChannel.Type, channel.Type)
	assert.Equal(t, expectedChannel.ID, channel.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertChannelRepository_GetByTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertChannelRepository(sqlxDB)

	now := time.Now()
	configJSON := []byte(`{}`)
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "type", "name", "config", "enabled", "created_at", "updated_at"}).
		AddRow("channel-1", 1, "webhook", "Webhook 1", configJSON, true, now, now).
		AddRow("channel-2", 1, "discord", "Discord 1", configJSON, true, now, now)

	mock.ExpectQuery(`SELECT id, tenant_id, type, name, config, enabled, created_at, updated_at FROM alert_channels WHERE tenant_id`).
		WithArgs(1).
		WillReturnRows(rows)

	channels, err := repo.GetByTenantID(1)
	assert.NoError(t, err)
	assert.Len(t, channels, 2)
	assert.Equal(t, "Webhook 1", channels[0].Name)
	assert.Equal(t, "Discord 1", channels[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertChannelRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertChannelRepository(sqlxDB)

	channel := &entities.AlertChannel{
		ID:       "channel-uuid",
		TenantID: 1,
		Type:     "email",
		Name:     "Updated Email",
		Config: entities.ChannelConfig{
			"to": "admin@example.com",
		},
		Enabled: false,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"updated_at"}).AddRow(now)

	mock.ExpectQuery(`UPDATE alert_channels SET`).
		WithArgs(channel.Type, channel.Name, sqlmock.AnyArg(), channel.Enabled, channel.ID).
		WillReturnRows(rows)

	err = repo.Update(channel)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertChannelRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertChannelRepository(sqlxDB)

	mock.ExpectExec(`DELETE FROM alert_channels WHERE id`).
		WithArgs("channel-uuid").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete("channel-uuid")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
