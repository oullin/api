package repository_test

import (
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
)

func TestUsersFindByPostgres(t *testing.T) {
	conn := newPostgresConnection(t, &database.User{})

	user := seedUser(t, conn, "Jane", "Doe", "janedoe")

	repo := repository.Users{DB: conn}

	if found := repo.FindBy("JaneDoe"); found == nil || found.ID != user.ID {
		t.Fatalf("expected to find user by case-insensitive username")
	}

	if found := repo.FindBy("missing"); found != nil {
		t.Fatalf("expected missing user lookup to return nil")
	}
}
