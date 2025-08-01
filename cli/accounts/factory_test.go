package accounts

import (
	clitest "github.com/oullin/cli/clitest"
	"testing"

	"github.com/oullin/database"
)

func TestMakeHandler(t *testing.T) {
	conn := clitest.MakeSQLiteConnection(t, &database.APIKey{})
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
