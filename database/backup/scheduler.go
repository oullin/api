package backup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/oullin/metal/env"
)

// CommandRunner defines an abstraction over exec.CommandContext so that
// backups can be tested without invoking external binaries.
type CommandRunner interface {
	Run(ctx context.Context, name string, args []string, env map[string]string) error
}

// ExecRunner executes commands using the local OS shell.
type ExecRunner struct{}

// Run executes the given command using exec.CommandContext. The process output
// is included in the returned error when the command fails.
func (ExecRunner) Run(ctx context.Context, name string, args []string, envVars map[string]string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), flattenEnv(envVars)...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %w: %s", name, err, string(output))
	}

	return nil
}

var (
	scheduleParser = cron.NewParser(
		cron.SecondOptional |
			cron.Minute |
			cron.Hour |
			cron.Dom |
			cron.Month |
			cron.Dow |
			cron.Descriptor,
	)
)

// Scheduler manages the lifecycle of the database backup cron routine.
type Scheduler struct {
	cron        *cron.Cron
	env         *env.Environment
	runner      CommandRunner
	logger      *slog.Logger
	now         func() time.Time
	jobTimeout  time.Duration
	started     bool
	startStopMu sync.Mutex
	entryID     cron.EntryID
}

// Option configures the scheduler.
type Option func(*Scheduler)

// WithCommandRunner allows providing a custom command runner (useful for tests).
func WithCommandRunner(runner CommandRunner) Option {
	return func(s *Scheduler) {
		if runner != nil {
			s.runner = runner
		}
	}
}

// WithLogger overrides the scheduler logger.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Scheduler) {
		if logger != nil {
			s.logger = logger
		}
	}
}

// WithNow allows tests to control the timestamp used for backup filenames.
func WithNow(now func() time.Time) Option {
	return func(s *Scheduler) {
		if now != nil {
			s.now = now
		}
	}
}

// WithJobTimeout configures a timeout applied to each backup execution.
func WithJobTimeout(timeout time.Duration) Option {
	return func(s *Scheduler) {
		if timeout > 0 {
			s.jobTimeout = timeout
		}
	}
}

// WithCron allows injecting a custom cron engine.
func WithCron(c *cron.Cron) Option {
	return func(s *Scheduler) {
		if c != nil {
			s.cron = c
		}
	}
}

// NewScheduler creates a backup scheduler using the provided environment.
func NewScheduler(environment *env.Environment, opts ...Option) (*Scheduler, error) {
	if environment == nil {
		return nil, errors.New("environment cannot be nil")
	}

	if _, err := scheduleParser.Parse(environment.Backup.Cron); err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	scheduler := &Scheduler{
		cron:       cron.New(cron.WithParser(scheduleParser)),
		env:        environment,
		runner:     ExecRunner{},
		logger:     slog.Default(),
		now:        time.Now,
		jobTimeout: 5 * time.Minute,
	}

	for _, opt := range opts {
		opt(scheduler)
	}

	if scheduler.cron == nil {
		scheduler.cron = cron.New(cron.WithParser(scheduleParser))
	}

	if scheduler.runner == nil {
		return nil, errors.New("command runner cannot be nil")
	}

	if scheduler.now == nil {
		scheduler.now = time.Now
	}

	return scheduler, nil
}

// Start schedules the backup routine and begins executing it according to the
// configured cron expression.
func (s *Scheduler) Start(ctx context.Context) error {
	if s == nil {
		return errors.New("scheduler is nil")
	}

	s.startStopMu.Lock()
	defer s.startStopMu.Unlock()

	if s.started {
		return errors.New("scheduler already started")
	}

	job := func() {
		jobCtx := ctx
		if jobCtx == nil {
			jobCtx = context.Background()
		}

		if s.jobTimeout > 0 {
			var cancel context.CancelFunc
			jobCtx, cancel = context.WithTimeout(jobCtx, s.jobTimeout)
			defer cancel()
		}

		if err := s.Run(jobCtx); err != nil {
			s.logger.Error("database backup failed", "error", err)
		}
	}

	entryID, err := s.cron.AddFunc(s.env.Backup.Cron, job)
	if err != nil {
		return fmt.Errorf("schedule backup job: %w", err)
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
func (s *Scheduler) Stop() {
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

// Run executes a database backup immediately using the scheduler configuration.
func (s *Scheduler) Run(ctx context.Context) error {
	if s == nil {
		return errors.New("scheduler is nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	backupDir := s.env.Backup.Dir
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}

	timestamp := s.now().UTC().Format("20060102T150405Z")
	fileName := fmt.Sprintf("backup-%s.sql", timestamp)
	filePath := filepath.Join(backupDir, fileName)

	args := []string{
		"--host", s.env.DB.Host,
		"--port", strconv.Itoa(s.env.DB.Port),
		"--username", s.env.DB.UserName,
		"--file", filePath,
		"--no-owner",
		"--no-privileges",
		s.env.DB.DatabaseName,
	}

	envVars := map[string]string{
		"PGPASSWORD": s.env.DB.UserPassword,
		"PGSSLMODE":  s.env.DB.SSLMode,
	}

	if err := s.runner.Run(ctx, "pg_dump", args, envVars); err != nil {
		return err
	}

	s.logger.Info("database backup created", "path", filePath)

	return nil
}

// flattenEnv converts a map of environment variables into a slice suitable for
// os/exec commands.
func flattenEnv(envVars map[string]string) []string {
	if len(envVars) == 0 {
		return nil
	}

	values := make([]string, 0, len(envVars))
	for key, value := range envVars {
		values = append(values, fmt.Sprintf("%s=%s", key, value))
	}

	return values
}
