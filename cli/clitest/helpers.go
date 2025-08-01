package clitest

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func MakeSQLiteConnection(t *testing.T, models ...interface{}) *database.Connection {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}
	conn := &database.Connection{}
	v := reflect.ValueOf(conn).Elem()
	field := v.FieldByName("driver")
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(db))
	return conn
}

func MakeTestEnv() *env.Environment {
	return &env.Environment{App: env.AppEnvironment{MasterKey: uuid.NewString()[:32]}}
}
