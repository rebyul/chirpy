-- name: DeleteUsers :many
delete from users
returning id;
