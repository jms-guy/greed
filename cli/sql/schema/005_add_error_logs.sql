-- +goose Up
CREATE TABLE error_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    command TEXT NOT NULL,
    error_message TEXT NOT NULL 
);

-- +goose Down
DROP TABLE error_logs;