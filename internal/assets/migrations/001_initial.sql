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

CREATE TABLE blocks (
    id SERIAL PRIMARY KEY,
    hash TEXT UNIQUE NOT NULL,
    height BIGINT NOT NULL UNIQUE
);

CREATE TABLE utxos (
    id SERIAL PRIMARY KEY,

    txid TEXT NOT NULL,
    vout INT NOT NULL,

    address_id INT REFERENCES addresses(id) ON DELETE CASCADE,
    block_id INT REFERENCES blocks(id) ON DELETE CASCADE
);

-- +migrate Down

DROP USER healthchecker;
DROP TABLE users;
DROP TABLE addresses;
DROP TABLE users_addresses;
DROP TABLE blocks;
DROP TABLE utxos;