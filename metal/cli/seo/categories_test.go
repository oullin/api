package seo_test

import (
	"slices"
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/internal/testutil/dbtest"
	"github.com/oullin/metal/cli/seo"
)

func TestCategoriesGenerateReturnsLowercaseNames(t *testing.T) {
	h := dbtest.NewTestsHelper(t, &database.Category{})

	h.SeedCategory("go", "GoLang", 1)
	h.SeedCategory("cli", "CLI Tools", 2)

	conn := h.Conn()

	categories, err := seo.NewCategories(conn).Generate()
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
