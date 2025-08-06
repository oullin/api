package accounts

import (
	"github.com/oullin/cli/clitest"
	"testing"

	"github.com/oullin/database"
)

func TestMakeHandler(t *testing.T) {
	conn := clitest.MakeTestConnection(t, &database.APIKey{})
	h, err := MakeHandler(conn, clitest.MakeTestEnv())

	if err != nil {
		t.Fatalf("make handler: %v", err)
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

func TestMakeHandlerInvalidKey(t *testing.T) {
	conn := clitest.MakeTestConnection(t)
	env := clitest.MakeTestEnv()
	env.App.MasterKey = "short"

	if _, err := MakeHandler(conn, env); err == nil {
		t.Fatalf("expected error")
	}
}
