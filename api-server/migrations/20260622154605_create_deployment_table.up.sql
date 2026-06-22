CREATE TYPE deployment_status AS ENUM (
    'pending',
    'building',
    'running',
    'failed',
    'stopped'
);

CREATE TABLE deployments(

    id BIGSERIAL PRIMARY KEY,

    project_id BIGINT NOT NULL REFERENCES projects(id),

    hostname TEXT UNIQUE NOT NULL,

    port INT UNIQUE,

    container_id TEXT,

    status deployment_status NOT NULL DEFAULT 'pending',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    retry_count INT NOT NULL DEFAULT 0
);