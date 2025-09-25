package repository_test

import (
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
)

func TestUsersFindBySQLite(t *testing.T) {
	conn, db := newSQLiteConnection(t)

	if err := db.AutoMigrate(&database.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}

	user := seedUser(t, conn, "Jane", "Doe", "janedoe")

	repo := repository.Users{DB: conn}

	if found := repo.FindBy("JaneDoe"); found == nil || found.ID != user.ID {
		t.Fatalf("expected to find user by case-insensitive username")
	}

	if found := repo.FindBy("missing"); found != nil {
		t.Fatalf("expected missing user lookup to return nil")
	}
}
