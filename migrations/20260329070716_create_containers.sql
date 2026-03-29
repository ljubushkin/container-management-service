-- +goose Up
CREATE TABLE containers (
    id UUID PRIMARY KEY,
    type_code TEXT NOT NULL REFERENCES container_types(code),
    warehouse_id TEXT NULL REFERENCES warehouses(id),
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_containers_type_code ON containers(type_code);
CREATE INDEX idx_containers_status ON containers(status);
CREATE INDEX idx_containers_warehouse_id ON containers(warehouse_id);
CREATE INDEX idx_containers_created_at ON containers(created_at);

-- +goose Down
DROP TABLE IF EXISTS containers;