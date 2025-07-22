-- +goose Up
DROP TABLE supported_currencies;

-- +goose Down
CREATE TABLE supported_currencies (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true
);

INSERT INTO supported_currencies (code, name)
VALUES
    ('CAD', 'Canadian Dollar'),
    ('USD', 'US Dollar'),
    ('EUR', 'Euro'),
    ('GBP', 'Great Britain Pound');