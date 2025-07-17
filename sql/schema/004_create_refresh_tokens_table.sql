-- +goose Up
CREATE TABLE refresh_tokens (
    token text PRIMARY KEY NOT NULL,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz NULL
);

-- +goose Down
DROP TABLE refresh_tokens;

