package clitest

import (
	"os/exec"
	"testing"
)

func TestNewTestEnv(t *testing.T) {
	env := NewTestEnv()

	if len(env.App.MasterKey) != 32 {
		t.Fatalf("expected master key length 32, got %d", len(env.App.MasterKey))
	}
}

func TestNewTestConnectionSkipsWithoutDocker(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}

	t.Run("skip", func(t *testing.T) {
		NewTestConnection(t)
	})
}
