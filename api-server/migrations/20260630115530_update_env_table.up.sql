ALTER TABLE project_env_vars
ALTER COLUMN encrypted_value
TYPE BYTEA
USING encrypted_value::BYTEA;