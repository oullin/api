CREATE TABLE api_keys_signatures (
	id BIGSERIAL PRIMARY KEY,
	uuid UUID UNIQUE NOT NULL,
	api_key_id BIGINT NOT NULL,
	signature BYTEA NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP DEFAULT NULL,

	CONSTRAINT unique_signature UNIQUE (signature, api_key_id, created_at),
	CONSTRAINT fk_api_key_id FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
);

CREATE INDEX idx_api_keys_signatures_signature_created_at ON api_keys_signatures(signature, created_at);
CREATE INDEX idx_api_keys_signatures_created_at ON api_keys_signatures(created_at);
CREATE INDEX idx_api_keys_signatures_deleted_at ON api_keys_signatures(deleted_at);
