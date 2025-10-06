package sqlseed

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oullin/database"
	"gorm.io/gorm"
)

func SeedFromFile(conn *database.Connection, filePath string) error {
	cleanedPath, err := validateFilePath(filePath)
	if err != nil {
		return err
	}

	statements, err := readSQLFile(cleanedPath)
	if err != nil {
		return err
	}

	if conn == nil {
		return errors.New("sqlseed: database connection is required")
	}

	return conn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(statements).Error; err != nil {
			return fmt.Errorf("sqlseed: executing SQL statements failed: %w", err)
		}

		return nil
	})
}

func validateFilePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("sqlseed: file path is required")
	}

	cleaned := filepath.Clean(path)
	if ext := strings.ToLower(filepath.Ext(cleaned)); ext != ".sql" {
		return "", fmt.Errorf("sqlseed: unsupported file extension %q", filepath.Ext(cleaned))
	}

	return cleaned, nil
}

func readSQLFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("sqlseed: read file: %w", err)
	}

	statements := strings.TrimSpace(string(bytes))
	if statements == "" {
		return "", fmt.Errorf("sqlseed: file %s is empty", path)
	}

	return statements, nil
}
