CREATE TABLE api_key_signatures (
	id BIGSERIAL PRIMARY KEY,
	uuid UUID UNIQUE NOT NULL,
	api_key_id BIGINT NOT NULL,
	signature BYTEA NOT NULL,
	max_tries SMALLINT NOT NULL DEFAULT 1 CHECK (max_tries > 0),
	current_tries SMALLINT NOT NULL DEFAULT 1 CHECK (current_tries > 0),
	expires_at TIMESTAMP DEFAULT NULL,
	expired_at TIMESTAMP DEFAULT NULL,
	origin TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP DEFAULT NULL,

	CONSTRAINT api_key_signatures_unique_signature UNIQUE (signature, api_key_id, created_at),
	CONSTRAINT api_key_signatures_fk_api_key_id FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
);

CREATE INDEX idx_api_key_signatures_signature_created_at ON api_key_signatures(signature, created_at);
CREATE INDEX idx_api_key_signatures_expires_at ON api_key_signatures(expires_at);
CREATE INDEX idx_api_key_signatures_expired_at ON api_key_signatures(expired_at);
CREATE INDEX idx_api_key_signatures_created_at ON api_key_signatures(created_at);
CREATE INDEX idx_api_key_signatures_deleted_at ON api_key_signatures(deleted_at);
