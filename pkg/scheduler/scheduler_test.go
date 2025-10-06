package scheduler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

type jobRecorder struct {
	mu    sync.Mutex
	calls int
	err   error
	hook  func()
}

func (j *jobRecorder) run(context.Context) error {
	j.mu.Lock()
	j.calls++
	hook := j.hook
	err := j.err
	j.mu.Unlock()

	if hook != nil {
		hook()
	}

	return err
}

func TestNewValidatesInput(t *testing.T) {
	t.Run("empty expression", func(t *testing.T) {
		if _, err := New("", func(context.Context) error { return nil }); err == nil {
			t.Fatalf("expected error when expression empty")
		}
	})

	t.Run("nil job", func(t *testing.T) {
		if _, err := New("@daily", nil); err == nil {
			t.Fatalf("expected error when job nil")
		}
	})

	t.Run("invalid expression", func(t *testing.T) {
		if _, err := New("not a cron", func(context.Context) error { return nil }); err == nil {
			t.Fatalf("expected error when expression invalid")
		}
	})
}

func TestRunInvokesJob(t *testing.T) {
	recorder := &jobRecorder{}

	scheduler, err := New("@weekly", recorder.run)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := scheduler.Run(context.Background()); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if recorder.calls != 1 {
		t.Fatalf("expected job to run once, ran %d", recorder.calls)
	}
}

func TestRunPropagatesErrors(t *testing.T) {
	scheduler, err := New("@hourly", func(context.Context) error { return errors.New("boom") })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := scheduler.Run(context.Background()); err == nil {
		t.Fatalf("expected error from job")
	}
}

func TestStartSchedulesJob(t *testing.T) {
	callCh := make(chan struct{}, 1)
	recorder := &jobRecorder{hook: func() {
		select {
		case callCh <- struct{}{}:
		default:
		}
	}}

	scheduler, err := New(
		"@every 1s",
		recorder.run,
		WithCron(cron.New(cron.WithParser(DefaultParser))),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := scheduler.Start(ctx); err != nil {
		t.Fatalf("start returned error: %v", err)
	}

	select {
	case <-callCh:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected job to run")
	}

	cancel()
	scheduler.Stop()
}

func TestStartWithNilContext(t *testing.T) {
	callCh := make(chan struct{}, 1)
	recorder := &jobRecorder{hook: func() {
		select {
		case callCh <- struct{}{}:
		default:
		}
	}}

	scheduler, err := New(
		"@every 1s",
		recorder.run,
		WithCron(cron.New(cron.WithParser(DefaultParser))),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := scheduler.Start(nil); err != nil {
		t.Fatalf("start returned error: %v", err)
	}

	select {
	case <-callCh:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected job to run")
	}

	scheduler.Stop()
}

func TestStartReturnsErrorWhenAlreadyStarted(t *testing.T) {
	scheduler, err := New("@daily", func(context.Context) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := scheduler.Start(context.Background()); err != nil {
		t.Fatalf("start returned error: %v", err)
	}

	t.Cleanup(scheduler.Stop)

	if err := scheduler.Start(context.Background()); err == nil {
		t.Fatalf("expected error when starting twice")
	}
}

func TestWithJobTimeout(t *testing.T) {
	scheduler, err := New("@daily", func(context.Context) error { return nil }, WithJobTimeout(time.Second))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if scheduler.jobTimeout != time.Second {
		t.Fatalf("expected job timeout to be set")
	}
}
