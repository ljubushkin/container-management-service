-- +goose Up
CREATE TABLE container_types (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE warehouses (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS warehouses;
DROP TABLE IF EXISTS container_types;