package scheduler

import (
	"context"
	"log"
	"sync"

	"github.com/robfig/cron/v3"
)

// Job represents a job that can be scheduled
type Job interface {
	Run(ctx context.Context) error
	Name() string
}

// Scheduler wraps the cron scheduler and provides job management
type Scheduler struct {
	cron    *cron.Cron
	jobs    map[string]cron.EntryID
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// New creates a new Scheduler instance
func New() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create cron with seconds precision (optional, for testing)
	// Use cron.WithSeconds() for testing with second-level precision
	c := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))

	return &Scheduler{
		cron:   c,
		jobs:   make(map[string]cron.EntryID),
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddJob adds a job to the scheduler with the given cron schedule
// schedule: cron expression (e.g., "* * * * *" for every minute)
// job: the job to run
func (s *Scheduler) AddJob(schedule string, job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Wrap the job to include context and error handling
	wrappedJob := func() {
		jobName := job.Name()
		log.Printf("[Scheduler] Starting job: %s", jobName)
		
		if err := job.Run(s.ctx); err != nil {
			log.Printf("[Scheduler] Job %s failed: %v", jobName, err)
		} else {
			log.Printf("[Scheduler] Job %s completed successfully", jobName)
		}
	}

	// Add job to cron
	entryID, err := s.cron.AddFunc(schedule, wrappedJob)
	if err != nil {
		return err
	}

	// Store the entry ID for this job
	s.jobs[job.Name()] = entryID
	log.Printf("[Scheduler] Added job: %s with schedule: %s", job.Name(), schedule)

	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	log.Println("[Scheduler] Starting cron scheduler...")
	s.cron.Start()
	log.Println("[Scheduler] Cron scheduler started successfully")
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	log.Println("[Scheduler] Stopping cron scheduler...")
	s.cancel() // Cancel context for all running jobs
	
	// Get context for stopping cron
	ctx := s.cron.Stop()
	<-ctx.Done() // Wait for all running jobs to complete
	
	log.Println("[Scheduler] Cron scheduler stopped successfully")
}

// GetJobs returns the list of registered job names
func (s *Scheduler) GetJobs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobNames := make([]string, 0, len(s.jobs))
	for name := range s.jobs {
		jobNames = append(jobNames, name)
	}
	return jobNames
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.jobs[jobName]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, jobName)
		log.Printf("[Scheduler] Removed job: %s", jobName)
	}
}
