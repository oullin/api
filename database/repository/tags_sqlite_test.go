package repository_test

import (
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
)

func TestTagsFindOrCreateSQLite(t *testing.T) {
	conn, db := newSQLiteConnection(t)

	if err := db.AutoMigrate(&database.Tag{}); err != nil {
		t.Fatalf("migrate tags: %v", err)
	}

	repo := repository.Tags{DB: conn}

	first, err := repo.FindOrCreate("golang")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}

	if first.Name != "Golang" {
		t.Fatalf("expected display name to be title case, got %q", first.Name)
	}

	second, err := repo.FindOrCreate("GOLANG")
	if err != nil {
		t.Fatalf("find tag: %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("expected idempotent lookup to reuse existing record")
	}

	if repo.FindBy("missing") != nil {
		t.Fatalf("expected missing lookup to return nil")
	}
}
