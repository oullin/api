DROP INDEX IF EXISTS idx_categories_sort;
ALTER TABLE categories
    DROP COLUMN IF EXISTS sort;
