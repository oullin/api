ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS sort INT;

CREATE INDEX IF NOT EXISTS idx_categories_sort ON categories (sort, name);
