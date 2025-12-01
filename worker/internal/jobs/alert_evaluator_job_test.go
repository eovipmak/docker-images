package jobs

import (
	"database/sql"
	"testing"

	"github.com/eovipmak/v-insight/shared/domain/entities"
)

// Note: These are placeholder tests. Full integration testing would require
// a test database instance or mocking framework.

func TestAlertEvaluatorJob_Name(t *testing.T) {
	job := &AlertEvaluatorJob{}
	
	expected := "AlertEvaluatorJob"
	if job.Name() != expected {
		t.Errorf("Expected job name %s, got %s", expected, job.Name())
	}
}

func TestEvaluateRule_Down(t *testing.T) {
	job := &AlertEvaluatorJob{}

	check := &entities.MonitorCheck{
		Success: false,
	}

	rule := &entities.AlertRule{
		TriggerType: "down",
	}

	triggered, value := job.evaluateRule(check, rule)
	if !triggered {
		t.Error("Expected down rule to be triggered when monitor is down")
	}

	if value == "" {
		t.Error("Expected trigger value to be set")
	}
}

func TestEvaluateRule_Down_NotTriggered(t *testing.T) {
	job := &AlertEvaluatorJob{}

	check := &entities.MonitorCheck{
		Success: true,
	}

	rule := &entities.AlertRule{
		TriggerType: "down",
	}

	triggered, _ := job.evaluateRule(check, rule)
	if triggered {
		t.Error("Expected down rule not to be triggered when monitor is up")
	}
}

func TestEvaluateRule_SlowResponse_Triggered(t *testing.T) {
	job := &AlertEvaluatorJob{}

	check := &entities.MonitorCheck{
		Success: true,
		ResponseTimeMs: sql.NullInt64{
			Int64: 5000,
			Valid: true,
		},
	}

	rule := &entities.AlertRule{
		TriggerType:    "slow_response",
		ThresholdValue: 3000,
	}

	triggered, value := job.evaluateRule(check, rule)
	if !triggered {
		t.Error("Expected slow_response rule to be triggered when response time exceeds threshold")
	}

	if value == "" {
		t.Error("Expected trigger value to be set")
	}
}

func TestEvaluateRule_SlowResponse_NotTriggered(t *testing.T) {
	job := &AlertEvaluatorJob{}

	check := &entities.MonitorCheck{
		Success: true,
		ResponseTimeMs: sql.NullInt64{
			Int64: 100,
			Valid: true,
		},
	}

	rule := &entities.AlertRule{
		TriggerType:    "slow_response",
		ThresholdValue: 3000,
	}

	triggered, _ := job.evaluateRule(check, rule)
	if triggered {
		t.Error("Expected slow_response rule not to be triggered when response time is below threshold")
	}
}
