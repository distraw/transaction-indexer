-- +migrate Up

CREATE USER healthchecker;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password BYTEA NOT NULL
);

CREATE TABLE addresses (
    id SERIAL PRIMARY KEY,
    addr TEXT NOT NULL UNIQUE
);

CREATE TABLE users_addresses (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    address_id INT REFERENCES addresses(id) ON DELETE CASCADE,
    PRIMARY KEY(user_id, address_id)
);

-- +migrate Down

DROP USER heathchecker;
DROP TABLE users;