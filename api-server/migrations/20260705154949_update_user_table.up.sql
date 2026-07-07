ALTER TABLE users
    ADD COLUMN password_hash BYTEA UNIQUE NOT NULL,
    ADD COLUMN profile_pic TEXT;