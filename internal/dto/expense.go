package dto

type CreateExpenseRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0" binding:"required"`
	Currency    string  `json:"currency" validate:"required,len=3,currency" binding:"required"`
	Category    string  `json:"category" validate:"required,oneof=travel meals 'office supplies'" binding:"required"`
	Description string  `json:"description" validate:"omitempty,max=500"`
	Receipt     string  `json:"receipt" validate:"omitempty,max=500"`
}

type UpdateExpenseRequest struct {
	Amount      *float64 `json:"amount,omitempty" validate:"omitempty,gt=0"`
	Currency    string   `json:"currency,omitempty" validate:"omitempty,len=3,currency"`
	Category    string   `json:"category,omitempty" validate:"omitempty,oneof=travel meals 'office supplies'"`
	Description string   `json:"description,omitempty" validate:"omitempty,max=500"`
	Receipt     string   `json:"receipt,omitempty" validate:"omitempty,max=500"`
	Status      string   `json:"status,omitempty" validate:"omitempty,oneof=pending approved rejected"`
}

type ExpenseResponse struct {
	ID          uint    `json:"id"`
	UserID      uint    `json:"user_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Receipt     string  `json:"receipt"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ExpenseListResponse struct {
	Expenses []ExpenseResponse `json:"expenses"`
	Page     int               `json:"page"`
	PerPage  int               `json:"per_page"`
	Total    int64             `json:"total"`
}

type ExpenseFilter struct {
	Category string `form:"category"`
	Status   string `form:"status"`
	Page     int    `form:"page,default=1"`
	PerPage  int    `form:"per_page,default=10"`
}
