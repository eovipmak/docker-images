package executor

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockTask is a mock task for testing
type MockTask struct {
	name       string
	shouldFail bool
	delay      time.Duration
}

func (m *MockTask) Name() string {
	return m.name
}

func (m *MockTask) Execute(ctx context.Context) error {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.shouldFail {
		return errors.New("task failed")
	}
	return nil
}

func TestExecutor_New(t *testing.T) {
	exec := New(DefaultConfig())
	if exec == nil {
		t.Fatal("Expected executor to be created, got nil")
	}
}

func TestExecutor_StartStop(t *testing.T) {
	exec := New(DefaultConfig())
	exec.Start()
	exec.Stop()
}

func TestExecutor_SubmitTask(t *testing.T) {
	exec := New(Config{
		WorkerCount: 2,
		QueueSize:   10,
	})
	exec.Start()
	defer exec.Stop()

	task := &MockTask{name: "test-task"}
	err := exec.Submit(task)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Wait for result
	select {
	case result := <-exec.Results():
		if result.TaskName != "test-task" {
			t.Fatalf("Expected task name 'test-task', got '%s'", result.TaskName)
		}
		if !result.Success {
			t.Fatalf("Expected task to succeed, got failure: %v", result.Error)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for task result")
	}
}

func TestExecutor_SubmitFailingTask(t *testing.T) {
	exec := New(Config{
		WorkerCount: 2,
		RetryCount:  2,
		RetryDelay:  100 * time.Millisecond,
		QueueSize:   10,
	})
	exec.Start()
	defer exec.Stop()

	task := &MockTask{name: "failing-task", shouldFail: true}
	err := exec.Submit(task)
	if err != nil {
		t.Fatalf("Expected no error submitting task, got: %v", err)
	}

	// Wait for result
	select {
	case result := <-exec.Results():
		if result.TaskName != "failing-task" {
			t.Fatalf("Expected task name 'failing-task', got '%s'", result.TaskName)
		}
		if result.Success {
			t.Fatal("Expected task to fail, got success")
		}
		if result.Error == nil {
			t.Fatal("Expected error, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for task result")
	}
}

func TestExecutor_ProcessMultipleTasks(t *testing.T) {
	exec := New(Config{
		WorkerCount: 3,
		QueueSize:   10,
	})
	exec.Start()
	defer exec.Stop()

	tasks := []Task{
		&MockTask{name: "task-1"},
		&MockTask{name: "task-2"},
		&MockTask{name: "task-3"},
	}

	for _, task := range tasks {
		err := exec.Submit(task)
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}
	}

	// Collect results
	resultsMap := make(map[string]bool)
	for i := 0; i < len(tasks); i++ {
		select {
		case result := <-exec.Results():
			resultsMap[result.TaskName] = result.Success
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for task results")
		}
	}

	if len(resultsMap) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(resultsMap))
	}

	for _, success := range resultsMap {
		if !success {
			t.Fatal("Expected all tasks to succeed")
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.WorkerCount <= 0 {
		t.Fatal("Expected positive worker count")
	}
	if cfg.QueueSize <= 0 {
		t.Fatal("Expected positive queue size")
	}
}
