package backup

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/oullin/metal/env"
)

type fakeRunner struct {
	mu     sync.Mutex
	calls  []runnerCall
	runErr error
	onRun  func()
}

type runnerCall struct {
	name string
	args []string
	env  map[string]string
}

func (f *fakeRunner) Run(_ context.Context, name string, args []string, envVars map[string]string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	call := runnerCall{name: name, args: append([]string(nil), args...), env: map[string]string{}}
	for k, v := range envVars {
		call.env[k] = v
	}

	f.calls = append(f.calls, call)

	if f.onRun != nil {
		f.onRun()
	}

	return f.runErr
}

func TestNewSchedulerValidatesInput(t *testing.T) {
	t.Run("nil environment", func(t *testing.T) {
		if _, err := NewScheduler(nil); err == nil {
			t.Fatalf("expected error when environment is nil")
		}
	})

	t.Run("invalid cron", func(t *testing.T) {
		env := &env.Environment{Backup: env.BackupEnvironment{Cron: "not-a-cron", Dir: t.TempDir()}}

		if _, err := NewScheduler(env); err == nil {
			t.Fatalf("expected cron validation error")
		}
	})
}

func TestSchedulerRunInvokesCommandRunner(t *testing.T) {
	tmpDir := t.TempDir()
	now := func() time.Time { return time.Date(2024, time.May, 1, 3, 4, 5, 0, time.UTC) }
	runner := &fakeRunner{}

	environment := &env.Environment{
		DB: env.DBEnvironment{
			UserName:     "usernamefoo",
			UserPassword: "passwordfoo",
			DatabaseName: "dbnamefoo",
			Port:         5432,
			Host:         "db.example.com",
			SSLMode:      "require",
			TimeZone:     "UTC",
		},
		Backup: env.BackupEnvironment{
			Cron: "@daily",
			Dir:  tmpDir,
		},
	}

	scheduler, err := NewScheduler(
		environment,
		WithCommandRunner(runner),
		WithNow(now),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)
	if err != nil {
		t.Fatalf("unexpected error creating scheduler: %v", err)
	}

	if err := scheduler.Run(context.Background()); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(runner.calls))
	}

	call := runner.calls[0]

	expectedFile := filepath.Join(tmpDir, "backup-20240501T030405Z.sql")
	expectedArgs := []string{
		"--host", "db.example.com",
		"--port", "5432",
		"--username", "usernamefoo",
		"--file", expectedFile,
		"--no-owner",
		"--no-privileges",
		"dbnamefoo",
	}

	if call.name != "pg_dump" {
		t.Fatalf("unexpected command name: %s", call.name)
	}

	if len(call.args) != len(expectedArgs) {
		t.Fatalf("unexpected number of args: %v", call.args)
	}

	for i, arg := range expectedArgs {
		if call.args[i] != arg {
			t.Fatalf("arg[%d] expected %q got %q", i, arg, call.args[i])
		}
	}

	if call.env["PGPASSWORD"] != "passwordfoo" {
		t.Fatalf("missing password env var")
	}

	if call.env["PGSSLMODE"] != "require" {
		t.Fatalf("missing sslmode env var")
	}
}

func TestSchedulerRunPropagatesErrors(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &fakeRunner{runErr: errors.New("boom")}

	environment := &env.Environment{
		DB: env.DBEnvironment{
			UserName:     "usernamefoo",
			UserPassword: "passwordfoo",
			DatabaseName: "dbnamefoo",
			Port:         5432,
			Host:         "db.example.com",
			SSLMode:      "require",
		},
		Backup: env.BackupEnvironment{Cron: "@weekly", Dir: tmpDir},
	}

	scheduler, err := NewScheduler(environment, WithCommandRunner(runner))
	if err != nil {
		t.Fatalf("unexpected error creating scheduler: %v", err)
	}

	if err := scheduler.Run(context.Background()); err == nil {
		t.Fatalf("expected error from runner")
	}
}

func TestSchedulerStartSchedulesJob(t *testing.T) {
	tmpDir := t.TempDir()
	callCh := make(chan struct{}, 1)

	runner := &fakeRunner{onRun: func() {
		select {
		case callCh <- struct{}{}:
		default:
		}
	}}

	environment := &env.Environment{
		DB: env.DBEnvironment{
			UserName:     "usernamefoo",
			UserPassword: "passwordfoo",
			DatabaseName: "dbnamefoo",
			Port:         5432,
			Host:         "db.example.com",
			SSLMode:      "require",
		},
		Backup: env.BackupEnvironment{Cron: "@every 1s", Dir: tmpDir},
	}

	customCron := cron.New(cron.WithParser(scheduleParser))
	scheduler, err := NewScheduler(
		environment,
		WithCommandRunner(runner),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		WithCron(customCron),
	)
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := scheduler.Start(ctx); err != nil {
		t.Fatalf("start returned error: %v", err)
	}

	select {
	case <-callCh:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected backup to run at least once")
	}

	cancel()
	scheduler.Stop()
}

func TestSchedulerStartWithNilContext(t *testing.T) {
	tmpDir := t.TempDir()
	callCh := make(chan struct{}, 1)

	runner := &fakeRunner{onRun: func() {
		select {
		case callCh <- struct{}{}:
		default:
		}
	}}

	environment := &env.Environment{
		DB:     env.DBEnvironment{UserName: "user", UserPassword: "pass", DatabaseName: "db", Port: 5432, Host: "host", SSLMode: "require"},
		Backup: env.BackupEnvironment{Cron: "@every 1s", Dir: tmpDir},
	}

	scheduler, err := NewScheduler(
		environment,
		WithCommandRunner(runner),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		WithCron(cron.New(cron.WithParser(scheduleParser))),
	)
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}

	if err := scheduler.Start(nil); err != nil {
		t.Fatalf("start returned error: %v", err)
	}

	select {
	case <-callCh:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected backup to run at least once")
	}

	scheduler.Stop()
}

func TestSchedulerStartReturnsErrorWhenAlreadyStarted(t *testing.T) {
	tmpDir := t.TempDir()

	environment := &env.Environment{
		DB:     env.DBEnvironment{UserName: "user", UserPassword: "pass", DatabaseName: "db", Port: 5432, Host: "host", SSLMode: "require"},
		Backup: env.BackupEnvironment{Cron: "@daily", Dir: tmpDir},
	}

	scheduler, err := NewScheduler(environment)
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}

	if err := scheduler.Start(context.Background()); err != nil {
		t.Fatalf("start returned error: %v", err)
	}

	t.Cleanup(scheduler.Stop)

	if err := scheduler.Start(context.Background()); err == nil {
		t.Fatalf("expected error when starting twice")
	}
}

func TestWithJobTimeoutOption(t *testing.T) {
	env := &env.Environment{
		DB:     env.DBEnvironment{Host: "db", Port: 5432, UserName: "user", UserPassword: "pass", DatabaseName: "db", SSLMode: "disable"},
		Backup: env.BackupEnvironment{Cron: "@daily", Dir: t.TempDir()},
	}

	scheduler, err := NewScheduler(env, WithJobTimeout(time.Second))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if scheduler.jobTimeout != time.Second {
		t.Fatalf("expected job timeout to be set, got %v", scheduler.jobTimeout)
	}
}

func TestFlattenEnv(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		if values := flattenEnv(nil); values != nil {
			t.Fatalf("expected nil slice for empty map")
		}
	})

	t.Run("populated map", func(t *testing.T) {
		envVars := map[string]string{"A": "1", "B": "2"}
		values := flattenEnv(envVars)

		if len(values) != len(envVars) {
			t.Fatalf("expected %d values, got %d", len(envVars), len(values))
		}

		seen := make(map[string]struct{})
		for _, value := range values {
			seen[value] = struct{}{}
		}

		if _, ok := seen["A=1"]; !ok {
			t.Fatalf("missing formatted A entry")
		}

		if _, ok := seen["B=2"]; !ok {
			t.Fatalf("missing formatted B entry")
		}
	})
}
