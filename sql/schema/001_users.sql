-- +goose Up
create table users(
  id uuid,
  created_at timestamp,
  updated_at timestamp,
  email text,
  PRIMARY KEY(id)
);

-- +goose Down
drop table users;
