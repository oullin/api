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

const storageSQLDir = "storage/sql"

func validateFilePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("sqlseed: file path is required")
	}

	if filepath.IsAbs(path) {
		return "", errors.New("sqlseed: absolute file paths are not supported")
	}

	cleanedInput := filepath.Clean(path)
	if ext := strings.ToLower(filepath.Ext(cleanedInput)); ext != ".sql" {
		return "", fmt.Errorf("sqlseed: unsupported file extension %q", filepath.Ext(cleanedInput))
	}

	base := filepath.Clean(storageSQLDir)
	baseWithSep := base + string(os.PathSeparator)

	var resolved string
	if strings.HasPrefix(cleanedInput, baseWithSep) {
		resolved = cleanedInput
	} else {
		resolved = filepath.Join(base, cleanedInput)
	}

	resolved = filepath.Clean(resolved)
	if !strings.HasPrefix(resolved, baseWithSep) {
		return "", fmt.Errorf("sqlseed: file path must be within %s", base)
	}

	return resolved, nil
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
