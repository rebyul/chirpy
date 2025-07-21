-- +gooseUp
ALTER TABLE users
    ADD COLUMN is_chirpy_red boolean NOT NULL DEFAULT FALSE;

-- +gooseDown
ALTER TABLE users
    DROP COLUMN is_chirpy_red;

