-- +goose Up
CREATE TABLE chirps (
    id uuid PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    body varchar(140) NOT NULL,
    user_id uuid NOT NULL
);

-- +goose Down
DROP TABLE chirps;

