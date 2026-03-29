-- +goose Up
INSERT INTO container_types (code, name) VALUES
    ('EURO_PALLET', 'Euro pallet'),
    ('BOX', 'Box');

INSERT INTO warehouses (id, name) VALUES
    ('w1', 'Main'),
    ('w2', 'Reserve');

-- +goose Down
DELETE FROM warehouses WHERE id IN ('w1', 'w2');
DELETE FROM container_types WHERE code IN ('EURO_PALLET', 'BOX');