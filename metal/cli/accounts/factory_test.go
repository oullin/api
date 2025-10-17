package accounts

import (
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/metal/cli/clitest"
)

func TestNewHandler(t *testing.T) {
	conn := clitest.NewTestConnection(t, &database.APIKey{})
	h, err := NewHandler(conn, clitest.NewTestEnv())

	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	if h.TokenHandler == nil || h.Tokens == nil {
		t.Fatalf("handler not properly initialized")
	}

	if err := h.CreateAccount("sampleaccount"); err != nil {
		t.Fatalf("create account: %v", err)
	}

	var key database.APIKey

	if err := conn.Sql().First(&key, "account_name = ?", "sampleaccount").Error; err != nil {
		t.Fatalf("key not saved: %v", err)
	}
}

func TestNewHandlerInvalidKey(t *testing.T) {
	conn := clitest.NewTestConnection(t)
	env := clitest.NewTestEnv()
	env.App.MasterKey = "short"

	if _, err := NewHandler(conn, env); err == nil {
		t.Fatalf("expected error")
	}
}
