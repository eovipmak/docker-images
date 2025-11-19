package executor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Task represents a unit of work to be executed
type Task interface {
	Execute(ctx context.Context) error
	Name() string
}

// Result represents the result of a task execution
type Result struct {
	TaskName  string
	Success   bool
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}

// Executor manages a pool of workers to execute tasks concurrently
type Executor struct {
	workerCount int
	retryCount  int
	retryDelay  time.Duration
	tasks       chan Task
	results     chan Result
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// Config holds executor configuration
type Config struct {
	WorkerCount int
	RetryCount  int
	RetryDelay  time.Duration
	QueueSize   int
}

// DefaultConfig returns default executor configuration
func DefaultConfig() Config {
	return Config{
		WorkerCount: 5,
		RetryCount:  3,
		RetryDelay:  time.Second * 2,
		QueueSize:   100,
	}
}

// New creates a new Executor with the given configuration
func New(cfg Config) *Executor {
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 5
	}
	if cfg.RetryCount < 0 {
		cfg.RetryCount = 0
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = time.Second
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}

	return &Executor{
		workerCount: cfg.WorkerCount,
		retryCount:  cfg.RetryCount,
		retryDelay:  cfg.RetryDelay,
		tasks:       make(chan Task, cfg.QueueSize),
		results:     make(chan Result, cfg.QueueSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the worker pool
func (e *Executor) Start() {
	log.Printf("[Executor] Starting worker pool with %d workers", e.workerCount)

	for i := 0; i < e.workerCount; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}

	log.Println("[Executor] Worker pool started successfully")
}

// worker is a worker goroutine that processes tasks from the queue
func (e *Executor) worker(id int) {
	defer e.wg.Done()

	log.Printf("[Executor] Worker %d started", id)

	for {
		select {
		case <-e.ctx.Done():
			log.Printf("[Executor] Worker %d stopping", id)
			return
		case task, ok := <-e.tasks:
			if !ok {
				log.Printf("[Executor] Worker %d: task channel closed", id)
				return
			}

			result := e.executeTask(task)
			
			// Send result (non-blocking)
			select {
			case e.results <- result:
			default:
				log.Printf("[Executor] Worker %d: result queue full, dropping result for %s", id, task.Name())
			}
		}
	}
}

// executeTask executes a task with retry logic
func (e *Executor) executeTask(task Task) Result {
	startTime := time.Now()
	var lastErr error

	for attempt := 0; attempt <= e.retryCount; attempt++ {
		if attempt > 0 {
			log.Printf("[Executor] Retrying task %s (attempt %d/%d)", task.Name(), attempt, e.retryCount)
			time.Sleep(e.retryDelay)
		}

		err := task.Execute(e.ctx)
		if err == nil {
			duration := time.Since(startTime)
			return Result{
				TaskName:  task.Name(),
				Success:   true,
				Error:     nil,
				Duration:  duration,
				Timestamp: time.Now(),
			}
		}

		lastErr = err
	}

	// All retries failed
	duration := time.Since(startTime)
	return Result{
		TaskName:  task.Name(),
		Success:   false,
		Error:     lastErr,
		Duration:  duration,
		Timestamp: time.Now(),
	}
}

// Submit submits a task for execution
func (e *Executor) Submit(task Task) error {
	select {
	case <-e.ctx.Done():
		return fmt.Errorf("executor is shutting down")
	case e.tasks <- task:
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// Results returns the results channel for reading task results
func (e *Executor) Results() <-chan Result {
	return e.results
}

// Stop gracefully stops the executor
func (e *Executor) Stop() {
	log.Println("[Executor] Stopping executor...")
	
	e.cancel()      // Signal all workers to stop
	close(e.tasks)  // Close task channel
	e.wg.Wait()     // Wait for all workers to finish
	close(e.results) // Close results channel
	
	log.Println("[Executor] Executor stopped successfully")
}

// ProcessTasksInParallel is a helper function to execute multiple tasks in parallel
func (e *Executor) ProcessTasksInParallel(tasks []Task) []Result {
	results := make([]Result, 0, len(tasks))
	
	// Submit all tasks
	for _, task := range tasks {
		if err := e.Submit(task); err != nil {
			log.Printf("[Executor] Failed to submit task %s: %v", task.Name(), err)
			results = append(results, Result{
				TaskName:  task.Name(),
				Success:   false,
				Error:     err,
				Timestamp: time.Now(),
			})
		}
	}

	// Collect results
	for i := 0; i < len(tasks); i++ {
		select {
		case result := <-e.results:
			results = append(results, result)
		case <-time.After(30 * time.Second):
			log.Println("[Executor] Timeout waiting for task results")
			break
		}
	}

	return results
}
