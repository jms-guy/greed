-- +goose Up
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

ALTER TABLE accounts
ADD CONSTRAINT accounts_currency_fk
FOREIGN KEY (currency)
REFERENCES supported_currencies(code);

-- +goose Down
ALTER TABLE accounts
DROP CONSTRAINT accounts_currency_fk;

DROP TABLE supported_currencies;