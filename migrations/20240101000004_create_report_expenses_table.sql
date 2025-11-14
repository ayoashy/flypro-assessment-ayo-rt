-- +goose Up
CREATE TABLE report_expenses (
    expense_report_id INTEGER NOT NULL,
    expense_id INTEGER NOT NULL,
    PRIMARY KEY (expense_report_id, expense_id),
    CONSTRAINT fk_report_expenses_report FOREIGN KEY (expense_report_id) REFERENCES expense_reports(id) ON DELETE CASCADE,
    CONSTRAINT fk_report_expenses_expense FOREIGN KEY (expense_id) REFERENCES expenses(id) ON DELETE CASCADE
);

CREATE INDEX idx_report_expenses_report_id ON report_expenses(expense_report_id);
CREATE INDEX idx_report_expenses_expense_id ON report_expenses(expense_id);

-- +goose Down
DROP INDEX IF EXISTS idx_report_expenses_expense_id;
DROP INDEX IF EXISTS idx_report_expenses_report_id;
DROP TABLE IF EXISTS report_expenses;

