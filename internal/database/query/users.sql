-- internal/database/query/users.sql

-- name: CreateUser :one
INSERT INTO users (
    phone_number,
    email,
    password_hash
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone_number = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;
