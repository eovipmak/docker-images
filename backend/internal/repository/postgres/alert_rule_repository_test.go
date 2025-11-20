package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertRuleRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertRuleRepository(sqlxDB)

	rule := &entities.AlertRule{
		TenantID:       1,
		MonitorID:      sql.NullString{String: "monitor-uuid", Valid: true},
		Name:           "Test Alert Rule",
		TriggerType:    "down",
		ThresholdValue: 3,
		Enabled:        true,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow("rule-uuid", now, now)

	mock.ExpectQuery(`INSERT INTO alert_rules`).
		WithArgs(rule.TenantID, rule.MonitorID, rule.Name, rule.TriggerType, rule.ThresholdValue, rule.Enabled).
		WillReturnRows(rows)

	err = repo.Create(rule)
	assert.NoError(t, err)
	assert.Equal(t, "rule-uuid", rule.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRuleRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertRuleRepository(sqlxDB)

	expectedRule := &entities.AlertRule{
		ID:             "rule-uuid",
		TenantID:       1,
		MonitorID:      sql.NullString{String: "monitor-uuid", Valid: true},
		Name:           "Test Alert Rule",
		TriggerType:    "down",
		ThresholdValue: 3,
		Enabled:        true,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "monitor_id", "name", "trigger_type", "threshold_value", "enabled", "created_at", "updated_at"}).
		AddRow(expectedRule.ID, expectedRule.TenantID, expectedRule.MonitorID, expectedRule.Name, expectedRule.TriggerType, expectedRule.ThresholdValue, expectedRule.Enabled, now, now)

	mock.ExpectQuery(`SELECT id, tenant_id, monitor_id, name, trigger_type, threshold_value, enabled, created_at, updated_at FROM alert_rules WHERE id`).
		WithArgs("rule-uuid").
		WillReturnRows(rows)

	rule, err := repo.GetByID("rule-uuid")
	assert.NoError(t, err)
	assert.Equal(t, expectedRule.Name, rule.Name)
	assert.Equal(t, expectedRule.TriggerType, rule.TriggerType)
	assert.Equal(t, expectedRule.ID, rule.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRuleRepository_GetByTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertRuleRepository(sqlxDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "monitor_id", "name", "trigger_type", "threshold_value", "enabled", "created_at", "updated_at"}).
		AddRow("rule-1", 1, sql.NullString{}, "Rule 1", "down", 3, true, now, now).
		AddRow("rule-2", 1, sql.NullString{}, "Rule 2", "ssl_expiry", 7, true, now, now)

	mock.ExpectQuery(`SELECT id, tenant_id, monitor_id, name, trigger_type, threshold_value, enabled, created_at, updated_at FROM alert_rules WHERE tenant_id`).
		WithArgs(1).
		WillReturnRows(rows)

	rules, err := repo.GetByTenantID(1)
	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "Rule 1", rules[0].Name)
	assert.Equal(t, "Rule 2", rules[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRuleRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertRuleRepository(sqlxDB)

	rule := &entities.AlertRule{
		ID:             "rule-uuid",
		TenantID:       1,
		MonitorID:      sql.NullString{String: "monitor-uuid", Valid: true},
		Name:           "Updated Rule",
		TriggerType:    "slow_response",
		ThresholdValue: 5000,
		Enabled:        false,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"updated_at"}).AddRow(now)

	mock.ExpectQuery(`UPDATE alert_rules SET`).
		WithArgs(rule.MonitorID, rule.Name, rule.TriggerType, rule.ThresholdValue, rule.Enabled, rule.ID).
		WillReturnRows(rows)

	err = repo.Update(rule)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRuleRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertRuleRepository(sqlxDB)

	mock.ExpectExec(`DELETE FROM alert_rules WHERE id`).
		WithArgs("rule-uuid").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete("rule-uuid")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRuleRepository_AttachChannels(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertRuleRepository(sqlxDB)

	ruleID := "rule-uuid"
	channelIDs := []string{"channel-1", "channel-2"}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO alert_rule_channels`).
		WithArgs(ruleID, channelIDs[0]).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO alert_rule_channels`).
		WithArgs(ruleID, channelIDs[1]).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.AttachChannels(ruleID, channelIDs)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertRuleRepository_GetChannelsByRuleID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewAlertRuleRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{"alert_channel_id"}).
		AddRow("channel-1").
		AddRow("channel-2")

	mock.ExpectQuery(`SELECT alert_channel_id FROM alert_rule_channels WHERE alert_rule_id`).
		WithArgs("rule-uuid").
		WillReturnRows(rows)

	channelIDs, err := repo.GetChannelsByRuleID("rule-uuid")
	assert.NoError(t, err)
	assert.Len(t, channelIDs, 2)
	assert.Equal(t, "channel-1", channelIDs[0])
	assert.Equal(t, "channel-2", channelIDs[1])
	assert.NoError(t, mock.ExpectationsWereMet())
}
