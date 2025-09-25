package seo

import (
	"slices"
	"testing"

	"github.com/oullin/database"
)

func TestCategoriesGenerateReturnsLowercaseNames(t *testing.T) {
	conn, _ := newPostgresConnection(t, &database.Category{})

	seedCategory(t, conn, "go", "GoLang")
	seedCategory(t, conn, "cli", "CLI Tools")

	categories, err := NewCategories(conn).Generate()
	if err != nil {
		t.Fatalf("generate err: %v", err)
	}

	if len(categories) != 2 {
		t.Fatalf("expected 2 categories got %d", len(categories))
	}

	slices.Sort(categories)

	if categories[0] != "cli tools" {
		t.Fatalf("expected lowercase name, got %q", categories[0])
	}

	if categories[1] != "golang" {
		t.Fatalf("expected lowercase name, got %q", categories[1])
	}
}
