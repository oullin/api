package accounts

import (
	"testing"

	"github.com/oullin/metal/cli/clitest"
	"github.com/oullin/database"
)

func setupAccountHandler(t *testing.T) *Handler {
	conn := clitest.MakeTestConnection(t, &database.APIKey{})
	h, err := MakeHandler(conn, clitest.MakeTestEnv())

	if err != nil {
		t.Fatalf("make handler: %v", err)
	}

	return h
}

func TestCreateReadSignature(t *testing.T) {
	h := setupAccountHandler(t)

	if err := h.CreateAccount("tester"); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := h.ReadAccount("tester"); err != nil {
		t.Fatalf("read: %v", err)
	}

	if err := h.CreateSignature("tester"); err != nil {
		t.Fatalf("signature: %v", err)
	}
}

func TestCreateAccountInvalid(t *testing.T) {
	h := setupAccountHandler(t)

	if err := h.CreateAccount("ab"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadAccountNotFound(t *testing.T) {
	h := setupAccountHandler(t)

	if err := h.ReadAccount("missing"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCreateSignatureNotFound(t *testing.T) {
	h := setupAccountHandler(t)

	if err := h.CreateSignature("missing"); err == nil {
		t.Fatalf("expected error")
	}
}
