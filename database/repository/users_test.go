package repository_test

import (
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/internal/testutil/dbtest"
)

func TestUsersFindByPostgres(t *testing.T) {
	h := dbtest.NewTestsHelper(t, &database.User{})

	user := h.SeedUser("Jane", "Doe", "janedoe")

	conn := h.Conn()

	repo := repository.Users{DB: conn}

	if found := repo.FindBy("JaneDoe"); found == nil || found.ID != user.ID {
		t.Fatalf("expected to find user by case-insensitive username")
	}

	if found := repo.FindBy("missing"); found != nil {
		t.Fatalf("expected missing user lookup to return nil")
	}
}
