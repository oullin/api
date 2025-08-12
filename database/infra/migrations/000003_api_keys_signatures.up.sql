CREATE TABLE api_keys_signatures (
	id BIGSERIAL PRIMARY KEY,
	uuid UUID UNIQUE NOT NULL,
	api_key_id BIGINT NOT NULL,
	signature BYTEA NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP DEFAULT NULL,

	CONSTRAINT uq_signature UNIQUE (signature, created_at),

	-- This constraint requires the 'api_keys' table to have a primary key column named 'id'.
	CONSTRAINT fk_api_key_id FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
);

CREATE INDEX idx_signature ON api_keys_signatures(signature);
CREATE INDEX idx_signature_created_at ON api_keys_signatures(created_at);
CREATE INDEX idx_signature_deleted_at ON api_keys_signatures(deleted_at);
