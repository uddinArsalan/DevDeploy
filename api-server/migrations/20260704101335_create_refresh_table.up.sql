BEGIN;

CREATE TABLE refresh_tokens(
    id BIGSERIAL PRIMARY KEY ,
    token_hash CHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    revoked_at TIMESTAMPTZ,
    replaced_by_token_id BIGINT REFERENCES refresh_tokens(id),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

COMMIT;