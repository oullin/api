ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS sort INT NOT NULL DEFAULT 0;

ALTER TABLE categories
    ALTER COLUMN sort DROP DEFAULT;

CREATE INDEX IF NOT EXISTS idx_categories_sort ON categories (sort, name);
