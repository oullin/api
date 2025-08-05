package clitest

import (
	"os/exec"
	"testing"
)

func TestMakeTestEnv(t *testing.T) {
	env := MakeTestEnv()
	if len(env.App.MasterKey) != 32 {
		t.Fatalf("expected master key length 32, got %d", len(env.App.MasterKey))
	}
}

func TestMakeTestConnectionSkipsWithoutDocker(t *testing.T) {
	if _, err := exec.LookPath("docker"); err == nil {
		t.Skip("docker available")
	}
	t.Run("skip", func(t *testing.T) {
		MakeTestConnection(t)
	})
}
