package seeds

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/oullin/database"
	"github.com/oullin/env"
)

func testConnection(t *testing.T, e *env.Environment) *database.Connection {
	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	conn := &database.Connection{}
	rv := reflect.ValueOf(conn).Elem()

	driverField := rv.FieldByName("driver")
	reflect.NewAt(driverField.Type(), unsafe.Pointer(driverField.UnsafeAddr())).Elem().Set(reflect.ValueOf(db))

	nameField := rv.FieldByName("driverName")
	reflect.NewAt(nameField.Type(), unsafe.Pointer(nameField.UnsafeAddr())).Elem().SetString("sqlite")

	envField := rv.FieldByName("env")
	reflect.NewAt(envField.Type(), unsafe.Pointer(envField.UnsafeAddr())).Elem().Set(reflect.ValueOf(e))

	err = db.AutoMigrate(
		&database.User{},
		&database.Post{},
		&database.Category{},
		&database.PostCategory{},
		&database.Tag{},
		&database.PostTag{},
		&database.PostView{},
		&database.Comment{},
		&database.Like{},
		&database.Newsletter{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return conn
}

func setupSeeder(t *testing.T) *Seeder {
	e := &env.Environment{App: env.AppEnvironment{Type: "local"}}
	conn := testConnection(t, e)
	return MakeSeeder(conn, e)
}

func TestSeederWorkflow(t *testing.T) {
	seeder := setupSeeder(t)

	if err := seeder.TruncateDB(); err != nil {
		t.Fatalf("truncate err: %v", err)
	}

	userA, userB := seeder.SeedUsers()
	posts := seeder.SeedPosts(userA, userB)
	categories := seeder.SeedCategories()
	tags := seeder.SeedTags()
	seeder.SeedComments(posts...)
	seeder.SeedLikes(posts...)
	seeder.SeedPostsCategories(categories, posts)
	seeder.SeedPostTags(tags, posts)
	seeder.SeedPostViews(posts, userA, userB)
	if err := seeder.SeedNewsLetters(); err != nil {
		t.Fatalf("newsletter err: %v", err)
	}

	var count int64

	seeder.dbConn.Sql().Model(&database.User{}).Count(&count)
	if count != 2 {
		t.Fatalf("expected 2 users got %d", count)
	}

	seeder.dbConn.Sql().Model(&database.Post{}).Count(&count)
	if count != 2 {
		t.Fatalf("expected 2 posts got %d", count)
	}

	seeder.dbConn.Sql().Model(&database.Category{}).Count(&count)
	if count == 0 {
		t.Fatalf("categories not seeded")
	}
}

func TestSeederEmptyMethods(t *testing.T) {
	seeder := setupSeeder(t)

	seeder.SeedPostsCategories(nil, nil)
	seeder.SeedPostTags(nil, nil)
	seeder.SeedPostViews(nil)

	var count int64

	seeder.dbConn.Sql().Model(&database.PostCategory{}).Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 post_categories")
	}

	seeder.dbConn.Sql().Model(&database.PostTag{}).Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 post_tags")
	}

	seeder.dbConn.Sql().Model(&database.PostView{}).Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 post_views")
	}
}
