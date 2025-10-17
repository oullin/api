package repository_test

import (
	"testing"

	"github.com/google/uuid"
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

func TestCategoriesGetOrdersBySort(t *testing.T) {
	conn := newPostgresConnection(t, &database.Category{})

	repo := repository.Categories{DB: conn}

	lowSort := 10
	low := database.Category{
		UUID: uuid.NewString(),
		Name: "Low",
		Slug: "low",
		Sort: &lowSort,
	}

	highSort := 20
	high := database.Category{
		UUID: uuid.NewString(),
		Name: "High",
		Slug: "high",
		Sort: &highSort,
	}

	if err := conn.Sql().Create(&high).Error; err != nil {
		t.Fatalf("create high sort: %v", err)
	}

	if err := conn.Sql().Create(&low).Error; err != nil {
		t.Fatalf("create low sort: %v", err)
	}

	items, err := repo.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(items))
	}

	if items[0].Slug != "low" || items[1].Slug != "high" {
		t.Fatalf("expected sort ordering, got %+v", items)
	}
}
