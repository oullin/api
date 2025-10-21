ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS sort INT NOT NULL DEFAULT 0;

ALTER TABLE categories
    ALTER COLUMN sort DROP DEFAULT;

ALTER TABLE post_categories
    DROP CONSTRAINT IF EXISTS post_categories_category_id_fkey;

ALTER TABLE categories
    RENAME TO categories_old;

CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL,
    name VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    sort INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

INSERT INTO categories (id, uuid, name, slug, description, sort, created_at, updated_at, deleted_at)
SELECT id, uuid, name, slug, description, sort, created_at, updated_at, deleted_at
FROM categories_old;

DROP TABLE categories_old;

CREATE INDEX IF NOT EXISTS idx_categories_sort ON categories (sort, name);

ALTER TABLE post_categories
    ADD CONSTRAINT post_categories_category_id_fkey FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE;

SELECT setval(pg_get_serial_sequence('categories', 'id'), COALESCE((SELECT MAX(id) FROM categories), 0));
