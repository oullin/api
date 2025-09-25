package repository_test

import (
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
)

func TestCategoriesFindByPostgres(t *testing.T) {
	conn := newPostgresConnection(t, &database.Category{})

	category := seedCategory(t, conn, "news", "News")

	repo := repository.Categories{DB: conn}

	if found := repo.FindBy("NEWS"); found == nil || found.ID != category.ID {
		t.Fatalf("expected to find category via case-insensitive slug")
	}

	if repo.FindBy("missing") != nil {
		t.Fatalf("expected missing category lookup to return nil")
	}
}
