package repository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
)

func TestCategoriesFindByPostgres(t *testing.T) {
	conn := newPostgresConnection(t, &database.Category{})

	category := seedCategory(t, conn, "news", "News", 1)

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

	low := database.Category{
		UUID: uuid.NewString(),
		Name: "Low",
		Slug: "low",
		Sort: 10,
	}

	high := database.Category{
		UUID: uuid.NewString(),
		Name: "High",
		Slug: "high",
		Sort: 20,
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

func TestCategoriesGetOrdersBySortAndName(t *testing.T) {
	conn := newPostgresConnection(t, &database.Category{})

	repo := repository.Categories{DB: conn}

	alpha := database.Category{
		UUID: uuid.NewString(),
		Name: "Alpha",
		Slug: "alpha",
		Sort: 10,
	}

	bravo := database.Category{
		UUID: uuid.NewString(),
		Name: "Bravo",
		Slug: "bravo",
		Sort: 10,
	}

	if err := conn.Sql().Create(&bravo).Error; err != nil {
		t.Fatalf("create bravo: %v", err)
	}

	if err := conn.Sql().Create(&alpha).Error; err != nil {
		t.Fatalf("create alpha: %v", err)
	}

	items, err := repo.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(items))
	}

	if items[0].Slug != "alpha" || items[1].Slug != "bravo" {
		t.Fatalf("expected secondary name ordering, got %+v", items)
	}
}

func TestCategoriesExistOrUpdatePreservesSortWhenZero(t *testing.T) {
	conn := newPostgresConnection(t, &database.Category{})

	repo := repository.Categories{DB: conn}

	original := database.Category{
		UUID: uuid.NewString(),
		Name: "Original",
		Slug: "original",
		Sort: 25,
	}

	if err := conn.Sql().Create(&original).Error; err != nil {
		t.Fatalf("create original: %v", err)
	}

	existed, err := repo.ExistOrUpdate(database.CategoriesAttrs{
		Slug: "original",
		Name: "Renamed",
		Sort: 0,
	})
	if err != nil {
		t.Fatalf("exist or update: %v", err)
	}

	if !existed {
		t.Fatalf("expected category to exist")
	}

	var updated database.Category
	if err := conn.Sql().Where("id = ?", original.ID).First(&updated).Error; err != nil {
		t.Fatalf("reload category: %v", err)
	}

	if updated.Sort != original.Sort {
		t.Fatalf("expected sort to remain %d, got %d", original.Sort, updated.Sort)
	}

	if updated.Name != "Renamed" {
		t.Fatalf("expected name to be updated, got %s", updated.Name)
	}
}

func TestCategoriesExistOrUpdatePreservesSortWhenNonZero(t *testing.T) {
	conn := newPostgresConnection(t, &database.Category{})

	repo := repository.Categories{DB: conn}

	original := database.Category{
		UUID: uuid.NewString(),
		Name: "Original",
		Slug: "original",
		Sort: 25,
	}

	if err := conn.Sql().Create(&original).Error; err != nil {
		t.Fatalf("create original: %v", err)
	}

	existed, err := repo.ExistOrUpdate(database.CategoriesAttrs{
		Slug: "original",
		Name: "Renamed",
		Sort: 30,
	})
	if err != nil {
		t.Fatalf("exist or update: %v", err)
	}

	if !existed {
		t.Fatalf("expected category to exist")
	}

	var updated database.Category
	if err := conn.Sql().Where("id = ?", original.ID).First(&updated).Error; err != nil {
		t.Fatalf("reload category: %v", err)
	}

	if updated.Sort != original.Sort {
		t.Fatalf("expected sort to remain %d, got %d", original.Sort, updated.Sort)
	}

	if updated.Name != "Renamed" {
		t.Fatalf("expected name to be updated, got %s", updated.Name)
	}
}
