package agenda

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Job defines the work that should be executed by the scheduler.
type Job func(context.Context) error

// DefaultParser provides standard cron parsing support including optional seconds
// and predefined descriptors such as "@daily".
var DefaultParser = cron.NewParser(
	cron.SecondOptional |
		cron.Minute |
		cron.Hour |
		cron.Dom |
		cron.Month |
		cron.Dow |
		cron.Descriptor,
)

// Engine orchestrates the execution of a job according to a cron expression.
type Engine struct {
	cron        *cron.Cron
	expression  string
	job         Job
	logger      *slog.Logger
	jobTimeout  time.Duration
	started     bool
	startStopMu sync.Mutex
	entryID     cron.EntryID
}

// EngineOption configures the scheduler engine.
type EngineOption func(*Engine)

// WithCron injects a preconfigured cron engine to use for scheduling.
func WithEngineCron(c *cron.Cron) EngineOption {
	return func(s *Engine) {
		if c != nil {
			s.cron = c
		}
	}
}

// WithLogger overrides the scheduler logger.
func WithEngineLogger(logger *slog.Logger) EngineOption {
	return func(s *Engine) {
		if logger != nil {
			s.logger = logger
		}
	}
}

// WithJobTimeout configures a timeout applied to each job execution.
func WithEngineJobTimeout(timeout time.Duration) EngineOption {
	return func(s *Engine) {
		if timeout > 0 {
			s.jobTimeout = timeout
		}
	}
}

// New creates a scheduler for the provided cron expression and job.
func New(expression string, job Job, opts ...EngineOption) (*Engine, error) {
	if expression == "" {
		return nil, errors.New("cron expression cannot be empty")
	}

	if job == nil {
		return nil, errors.New("job cannot be nil")
	}

	if _, err := DefaultParser.Parse(expression); err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	scheduler := &Engine{
		expression: expression,
		job:        job,
		logger:     slog.Default(),
		jobTimeout: 0,
	}

	for _, opt := range opts {
		opt(scheduler)
	}

	if scheduler.cron == nil {
		scheduler.cron = cron.New(cron.WithParser(DefaultParser))
	}

	return scheduler, nil
}

// Start schedules the job according to the configured cron expression.
func (s *Engine) Start(ctx context.Context) error {
	if s == nil {
		return errors.New("scheduler is nil")
	}

	s.startStopMu.Lock()
	defer s.startStopMu.Unlock()

	if s.started {
		return errors.New("scheduler already started")
	}

	job := func() {
		if err := s.Run(ctx); err != nil {
			s.logger.Error("scheduled job failed", "error", err)
		}
	}

	entryID, err := s.cron.AddFunc(s.expression, job)
	if err != nil {
		return fmt.Errorf("schedule job: %w", err)
	}

	s.entryID = entryID
	s.cron.Start()
	s.started = true

	if ctx != nil {
		go func() {
			<-ctx.Done()
			s.Stop()
		}()
	}

	return nil
}

// Stop halts the scheduler and waits for any running job to finish.
func (s *Engine) Stop() {
	if s == nil {
		return
	}

	s.startStopMu.Lock()
	if !s.started {
		s.startStopMu.Unlock()
		return
	}

	ctx := s.cron.Stop()
	s.started = false
	s.startStopMu.Unlock()

	<-ctx.Done()
}

// Run executes the job immediately using the scheduler configuration.
func (s *Engine) Run(ctx context.Context) error {
	if s == nil {
		return errors.New("scheduler is nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if s.jobTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.jobTimeout)
		defer cancel()
	}

	return s.job(ctx)
}
