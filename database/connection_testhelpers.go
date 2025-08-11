package database

import "gorm.io/gorm"

// NewConnectionFromGorm is intended for tests only.
func NewConnectionFromGorm(db *gorm.DB) *Connection {
	return &Connection{driver: db}
}
