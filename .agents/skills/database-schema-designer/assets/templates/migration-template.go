package migrations

import (
	"time"

	"gorm.io/gorm"
)

// Migration: YYYYMMDDHHMMSS_descriptive_name.go
// Description: [What this migration does]
// Author: [Your Name]
// Date: YYYY-MM-DD

// ============================================================================
// MODELS (for GORM Auto-Migration)
// ============================================================================

// Example model for auto-migration
type TableName struct {
	ID          uint           `gorm:"primaryKey;autoIncrement"`
	ColumnName  string         `gorm:"type:varchar(255);not null;index:idx_table_column"`
	ReferenceID *uint          `gorm:"index"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Foreign key relationship
	Reference *OtherTable `gorm:"foreignKey:ReferenceID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// TableName overrides the default table name
func (TableName) TableName() string {
	return "table_name"
}

// ============================================================================
// UP MIGRATION
// ============================================================================

// Up runs the forward migration
func Up_YYYYMMDDHHMMSS(db *gorm.DB) error {
	// Option 1: Using GORM Auto-Migration
	if err := db.AutoMigrate(&TableName{}); err != nil {
		return err
	}

	// Option 2: Using raw SQL for complex migrations
	sql := `
		-- Step 1: Create table with PostgreSQL specific features
		CREATE TABLE IF NOT EXISTS table_name (
			id BIGSERIAL PRIMARY KEY,
			column_name VARCHAR(255) NOT NULL,
			json_data JSONB,
			array_data TEXT[],
			uuid_field UUID DEFAULT gen_random_uuid(),
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		-- Step 2: Add indexes
		CREATE INDEX IF NOT EXISTS idx_table_column ON table_name(column_name);
		CREATE INDEX IF NOT EXISTS idx_json_data ON table_name USING GIN (json_data);

		-- Step 3: Add constraints
		ALTER TABLE table_name
			ADD CONSTRAINT uk_table_unique UNIQUE (column_name);

		-- Step 4: Create trigger for updated_at
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		CREATE TRIGGER update_table_name_updated_at
			BEFORE UPDATE ON table_name
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();

		-- Step 5: Add foreign keys
		ALTER TABLE table_name
			ADD CONSTRAINT fk_table_reference
			FOREIGN KEY (reference_id) REFERENCES other_table(id)
			ON DELETE CASCADE;
	`

	if err := db.Exec(sql).Error; err != nil {
		return err
	}

	// Step 6: Data migration (if needed)
	if err := db.Exec(`
		UPDATE table_name
		SET json_data = '{"migrated": true}'::jsonb
		WHERE json_data IS NULL
	`).Error; err != nil {
		return err
	}

	return nil
}

// ============================================================================
// DOWN MIGRATION
// ============================================================================

// Down runs the reverse migration
func Down_YYYYMMDDHHMMSS(db *gorm.DB) error {
	// Option 1: Using GORM
	if err := db.Migrator().DropTable(&TableName{}); err != nil {
		return err
	}

	// Option 2: Using raw SQL
	sql := `
		-- Drop trigger first
		DROP TRIGGER IF EXISTS update_table_name_updated_at ON table_name;
		DROP FUNCTION IF EXISTS update_updated_at_column();

		-- Drop constraints
		ALTER TABLE IF EXISTS table_name DROP CONSTRAINT IF EXISTS fk_table_reference;
		ALTER TABLE IF EXISTS table_name DROP CONSTRAINT IF EXISTS uk_table_unique;

		-- Drop indexes
		DROP INDEX IF EXISTS idx_json_data;
		DROP INDEX IF EXISTS idx_table_column;

		-- Drop table
		DROP TABLE IF EXISTS table_name;
	`

	return db.Exec(sql).Error
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate checks if the migration was successful
func Validate_YYYYMMDDHHMMSS(db *gorm.DB) error {
	// Check table exists
	var exists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'table_name'
		)
	`).Scan(&exists).Error

	if err != nil {
		return err
	}

	// Check indexes
	var indexCount int64
	err = db.Raw(`
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE schemaname = 'public'
		AND tablename = 'table_name'
	`).Scan(&indexCount).Error

	if err != nil {
		return err
	}

	// Check constraints
	var constraintCount int64
	err = db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.table_constraints
		WHERE table_schema = 'public'
		AND table_name = 'table_name'
	`).Scan(&constraintCount).Error

	return err
}

// ============================================================================
// MIGRATION RUNNER (Optional - for standalone execution)
// ============================================================================

// MigrationInfo provides metadata about this migration
type MigrationInfo struct {
	Version          string
	Description      string
	Author           string
	Date             time.Time
	EstimatedTime    string
	RequiresDowntime bool
	RollbackTested   bool
}

// GetInfo returns migration metadata
func GetInfo_YYYYMMDDHHMMSS() MigrationInfo {
	return MigrationInfo{
		Version:          "YYYYMMDDHHMMSS",
		Description:      "[What this migration does]",
		Author:           "[Your Name]",
		Date:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EstimatedTime:    "[X seconds on Y rows]",
		RequiresDowntime: false,
		RollbackTested:   false,
	}
}

// ============================================================================
// NOTES
// ============================================================================
/*
PostgreSQL Specific Features Used:
- BIGSERIAL for auto-incrementing IDs
- JSONB for JSON data with indexing
- TEXT[] for array types
- UUID with gen_random_uuid()
- TIMESTAMPTZ for timezone-aware timestamps
- GIN indexes for JSONB fields
- Triggers for automatic updated_at

GORM Features:
- AutoMigrate for schema synchronization
- Soft deletes with DeletedAt
- Foreign key constraints with cascading
- Custom table names
- Index creation via tags

Best Practices:
1. Always wrap migrations in transactions when possible
2. Test rollback procedures before production deployment
3. Consider table locking implications for large tables
4. Use CONCURRENTLY for index creation on production (outside transactions)
5. Monitor migration performance on staging environment first
*/
