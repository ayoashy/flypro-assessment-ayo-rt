-- +goose Up
CREATE TABLE expense_reports (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    title VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    total DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_expense_reports_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_expense_reports_user_id ON expense_reports(user_id);
CREATE INDEX idx_expense_reports_status ON expense_reports(status);
CREATE INDEX idx_expense_reports_deleted_at ON expense_reports(deleted_at);

-- +goose Down
DROP INDEX IF EXISTS idx_expense_reports_deleted_at;
DROP INDEX IF EXISTS idx_expense_reports_status;
DROP INDEX IF EXISTS idx_expense_reports_user_id;
DROP TABLE IF EXISTS expense_reports;

