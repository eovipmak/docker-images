package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// MockJob is a mock job for testing
type MockJob struct {
	name       string
	runCount   int32
	shouldFail bool
}

func (m *MockJob) Name() string {
	return m.name
}

func (m *MockJob) Run(ctx context.Context) error {
	atomic.AddInt32(&m.runCount, 1)
	if m.shouldFail {
		return context.DeadlineExceeded
	}
	return nil
}

func (m *MockJob) GetRunCount() int32 {
	return atomic.LoadInt32(&m.runCount)
}

func TestScheduler_New(t *testing.T) {
	sched := New()
	if sched == nil {
		t.Fatal("Expected scheduler to be created, got nil")
	}
	if sched.cron == nil {
		t.Fatal("Expected cron to be initialized")
	}
	if sched.jobs == nil {
		t.Fatal("Expected jobs map to be initialized")
	}
}

func TestScheduler_AddJob(t *testing.T) {
	sched := New()
	job := &MockJob{name: "test-job"}

	err := sched.AddJob("* * * * *", job)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	jobs := sched.GetJobs()
	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}
	if jobs[0] != "test-job" {
		t.Fatalf("Expected job name 'test-job', got '%s'", jobs[0])
	}
}

func TestScheduler_AddJobInvalidSchedule(t *testing.T) {
	sched := New()
	job := &MockJob{name: "test-job"}

	err := sched.AddJob("invalid schedule", job)
	if err == nil {
		t.Fatal("Expected error for invalid schedule, got nil")
	}
}

func TestScheduler_StartStop(t *testing.T) {
	sched := New()
	job := &MockJob{name: "test-job"}

	// Use every second for faster testing
	err := sched.AddJob("* * * * * *", job)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	sched.Start()
	
	// Wait for job to run at least once
	time.Sleep(2 * time.Second)
	
	sched.Stop()

	runCount := job.GetRunCount()
	if runCount < 1 {
		t.Fatalf("Expected job to run at least once, got %d runs", runCount)
	}
}

func TestScheduler_RemoveJob(t *testing.T) {
	sched := New()
	job := &MockJob{name: "test-job"}

	err := sched.AddJob("* * * * *", job)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	jobs := sched.GetJobs()
	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}

	sched.RemoveJob("test-job")

	jobs = sched.GetJobs()
	if len(jobs) != 0 {
		t.Fatalf("Expected 0 jobs after removal, got %d", len(jobs))
	}
}

func TestScheduler_MultipleJobs(t *testing.T) {
	sched := New()
	job1 := &MockJob{name: "job-1"}
	job2 := &MockJob{name: "job-2"}

	err := sched.AddJob("* * * * * *", job1)
	if err != nil {
		t.Fatalf("Failed to add job1: %v", err)
	}

	err = sched.AddJob("* * * * * *", job2)
	if err != nil {
		t.Fatalf("Failed to add job2: %v", err)
	}

	jobs := sched.GetJobs()
	if len(jobs) != 2 {
		t.Fatalf("Expected 2 jobs, got %d", len(jobs))
	}

	sched.Start()
	time.Sleep(2 * time.Second)
	sched.Stop()

	if job1.GetRunCount() < 1 {
		t.Fatalf("Expected job1 to run at least once, got %d runs", job1.GetRunCount())
	}
	if job2.GetRunCount() < 1 {
		t.Fatalf("Expected job2 to run at least once, got %d runs", job2.GetRunCount())
	}
}
