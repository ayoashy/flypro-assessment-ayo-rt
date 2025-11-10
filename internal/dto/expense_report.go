package dto

type CreateExpenseReportRequest struct {
	Title string `json:"title" validate:"required,min=3,max=200" binding:"required"`
}

type AddExpensesToReportRequest struct {
	ExpenseIDs []uint `json:"expense_ids" validate:"required,min=1,dive,gt=0" binding:"required"`
}

type ExpenseReportResponse struct {
	ID        uint             `json:"id"`
	UserID    uint             `json:"user_id"`
	Title     string           `json:"title"`
	Status    string           `json:"status"`
	Total     float64          `json:"total"`
	CreatedAt string           `json:"created_at"`
	UpdatedAt string           `json:"updated_at"`
	Expenses  []ExpenseResponse `json:"expenses"`
}

type ExpenseReportListResponse struct {
	Reports  []ExpenseReportResponse `json:"reports"`
	Page     int                     `json:"page"`
	PerPage  int                     `json:"per_page"`
	Total    int64                   `json:"total"`
}

type ReportFilter struct {
	Status  string `form:"status"`
	Page    int    `form:"page,default=1"`
	PerPage int    `form:"per_page,default=10"`
}