-- +goose Up
CREATE INDEX IF NOT EXISTS idx_expenses_created_at ON expenses(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_expense_reports_created_at ON expense_reports(created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_expense_reports_created_at;
DROP INDEX IF EXISTS idx_expenses_created_at;

