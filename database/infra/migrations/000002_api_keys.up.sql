CREATE TABLE api_keys (
	id BIGSERIAL PRIMARY KEY,
	uuid UUID UNIQUE NOT NULL,
	account_name VARCHAR(255) UNIQUE NOT NULL,
	public_key VARCHAR(255) UNIQUE NOT NULL,
	secret_key VARCHAR(255) UNIQUE NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP DEFAULT NULL,

	CONSTRAINT uq_account_keys UNIQUE (account_name, public_key, secret_key)
);

CREATE INDEX idx_account_name ON api_keys(account_name);
CREATE INDEX idx_public_key ON api_keys(public_key);
CREATE INDEX idx_secret_key ON api_keys(secret_key);
CREATE INDEX idx_deleted_at ON api_keys(deleted_at);
