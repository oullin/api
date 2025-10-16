package accounts

import (
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/metal/cli/clitest"
)

func setupAccountHandler(t *testing.T) *Handler {
	conn := clitest.NewTestConnection(t, &database.APIKey{})
	h, err := NewHandler(conn, clitest.NewTestEnv())

	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	return h
}

func TestCreateReadSignature(t *testing.T) {
	h := setupAccountHandler(t)

	if err := h.CreateAccount("tester"); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := h.ShowAccount("tester"); err != nil {
		t.Fatalf("read: %v", err)
	}
}

func TestCreateAccountInvalid(t *testing.T) {
	h := setupAccountHandler(t)

	if err := h.CreateAccount("ab"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestShowAccountNotFound(t *testing.T) {
	h := setupAccountHandler(t)

	if err := h.ShowAccount("missing"); err == nil {
		t.Fatalf("expected error")
	}
}
