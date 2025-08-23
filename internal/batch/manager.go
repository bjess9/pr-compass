package batch

import (
	"context"
	"sync"
)

// Job represents a unit of work to be processed by the batch manager
type Job[T any, R any] struct {
	Input    T
	ResultCh chan Result[R]
}

// Result represents the result of processing a job
type Result[R any] struct {
	Data  R
	Error error
}

// WorkerFunc defines the function signature for processing jobs
type WorkerFunc[T any, R any] func(ctx context.Context, input T) (R, error)

// Manager handles concurrent batch processing with a worker pool
type Manager[T any, R any] struct {
	workerCount int
	jobQueue    chan Job[T, R]
	workerFunc  WorkerFunc[T, R]
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	started     bool
	mu          sync.Mutex
}

// NewManager creates a new batch manager with the specified number of workers
func NewManager[T any, R any](workerCount int, workerFunc WorkerFunc[T, R]) *Manager[T, R] {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Manager[T, R]{
		workerCount: workerCount,
		jobQueue:    make(chan Job[T, R], workerCount*2), // Buffer for smooth processing
		workerFunc:  workerFunc,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins the worker pool processing
func (m *Manager[T, R]) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.started {
		return // Already started
	}
	
	m.started = true
	
	// Start worker goroutines
	for i := 0; i < m.workerCount; i++ {
		m.wg.Add(1)
		go m.worker()
	}
}

// Stop gracefully shuts down the manager and waits for all workers to complete
func (m *Manager[T, R]) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.started {
		return // Not started
	}
	
	// Cancel context to signal workers to stop
	m.cancel()
	
	// Close job queue to prevent new jobs
	close(m.jobQueue)
	
	// Wait for all workers to finish
	m.wg.Wait()
	
	m.started = false
}

// Submit adds a job to the processing queue and returns a channel to receive the result
func (m *Manager[T, R]) Submit(input T) <-chan Result[R] {
	resultCh := make(chan Result[R], 1)
	
	// Handle panic from sending to closed channel
	defer func() {
		if r := recover(); r != nil {
			// Channel was closed, return cancellation error
			go func() {
				resultCh <- Result[R]{Error: context.Canceled}
			}()
		}
	}()
	
	job := Job[T, R]{
		Input:    input,
		ResultCh: resultCh,
	}
	
	// Try to submit the job, handle context cancellation
	select {
	case m.jobQueue <- job:
		// Job submitted successfully
	case <-m.ctx.Done():
		// Manager is shutting down, return error immediately
		go func() {
			resultCh <- Result[R]{Error: m.ctx.Err()}
		}()
	}
	
	return resultCh
}

// ProcessBatch processes multiple inputs concurrently and returns results
func (m *Manager[T, R]) ProcessBatch(inputs []T) []Result[R] {
	if len(inputs) == 0 {
		return nil
	}
	
	// Ensure manager is started
	m.Start()
	
	// Submit all jobs and collect result channels
	resultChannels := make([]<-chan Result[R], len(inputs))
	for i, input := range inputs {
		resultChannels[i] = m.Submit(input)
	}
	
	// Collect results
	results := make([]Result[R], len(inputs))
	for i, resultCh := range resultChannels {
		results[i] = <-resultCh
	}
	
	return results
}

// ProcessBatchWithCallback processes inputs concurrently and calls callback for each result
func (m *Manager[T, R]) ProcessBatchWithCallback(inputs []T, callback func(int, Result[R])) {
	if len(inputs) == 0 {
		return
	}
	
	// Ensure manager is started
	m.Start()
	
	// Submit all jobs and collect result channels
	resultChannels := make([]<-chan Result[R], len(inputs))
	for i, input := range inputs {
		resultChannels[i] = m.Submit(input)
	}
	
	// Process results as they come in
	var wg sync.WaitGroup
	for i, resultCh := range resultChannels {
		wg.Add(1)
		go func(index int, ch <-chan Result[R]) {
			defer wg.Done()
			result := <-ch
			callback(index, result)
		}(i, resultCh)
	}
	
	wg.Wait()
}

// worker is the main worker goroutine that processes jobs from the queue
func (m *Manager[T, R]) worker() {
	defer m.wg.Done()
	
	for {
		select {
		case job, ok := <-m.jobQueue:
			if !ok {
				// Job queue closed, exit worker
				return
			}
			
			// Process the job
			result, err := m.workerFunc(m.ctx, job.Input)
			
			// Send result back
			select {
			case job.ResultCh <- Result[R]{Data: result, Error: err}:
				// Result sent successfully
			case <-m.ctx.Done():
				// Context cancelled, exit worker
				return
			}
			
		case <-m.ctx.Done():
			// Context cancelled, exit worker
			return
		}
	}
}

// IsRunning returns true if the manager is currently running
func (m *Manager[T, R]) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.started
}

// WorkerCount returns the number of workers in the pool
func (m *Manager[T, R]) WorkerCount() int {
	return m.workerCount
}
