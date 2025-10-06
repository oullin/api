package agenda

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// Scheduler manages the lifecycle of the database backup routine.
type Scheduler struct {
	env        *env.Environment
	runner     CommandRunner
	logger     *slog.Logger
	now        func() time.Time
	jobTimeout time.Duration
	cron       *cron.Cron
	engine     *Engine
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

	scheduler := &Scheduler{
		env:        environment,
		runner:     ExecRunner{},
		logger:     slog.Default(),
		now:        time.Now,
		jobTimeout: 0,
	}

	for _, opt := range opts {
		opt(scheduler)
	}

	if scheduler.runner == nil {
		return nil, errors.New("command runner cannot be nil")
	}

	if scheduler.now == nil {
		scheduler.now = time.Now
	}

	job := func(ctx context.Context) error {
		return scheduler.runBackup(ctx)
	}

	engineOpts := []EngineOption{
		WithEngineLogger(scheduler.logger),
		WithEngineJobTimeout(scheduler.jobTimeout),
	}

	if scheduler.cron != nil {
		engineOpts = append(engineOpts, WithEngineCron(scheduler.cron))
	}

	engine, err := New(environment.Backup.Cron, job, engineOpts...)
	if err != nil {
		return nil, err
	}

	scheduler.engine = engine

	return scheduler, nil
}

// Start schedules the backup routine and begins executing it according to the
// configured cron expression.
func (s *Scheduler) Start(ctx context.Context) error {
	if s == nil {
		return errors.New("scheduler is nil")
	}

	if s.engine == nil {
		return errors.New("internal scheduler is nil")
	}

	return s.engine.Start(ctx)
}

// Stop halts the scheduler and waits for any running job to finish.
func (s *Scheduler) Stop() {
	if s == nil {
		return
	}

	if s.engine == nil {
		return
	}

	s.engine.Stop()
}

// Run executes a database backup immediately using the scheduler configuration.
func (s *Scheduler) Run(ctx context.Context) error {
	if s == nil {
		return errors.New("scheduler is nil")
	}

	if s.engine == nil {
		return errors.New("internal scheduler is nil")
	}

	return s.engine.Run(ctx)
}

func (s *Scheduler) runBackup(ctx context.Context) error {
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
