package models

import (
	"time"

	"gorm.io/gorm"
)

type Expense struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Amount      float64        `json:"amount" gorm:"not null"`
	Currency    string         `json:"currency" gorm:"not null"`
	Category    string         `json:"category" gorm:"not null;index"`
	Description string         `json:"description"`
	Receipt     string         `json:"receipt"`
	Status      string         `json:"status" gorm:"default:'pending';index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
}

func (Expense) TableName() string {
	return "expenses"
}

const (
	ExpenseStatusPending  = "pending"
	ExpenseStatusApproved = "approved"
	ExpenseStatusRejected = "rejected"
)

const (
	CategoryTravel         = "travel"
	CategoryMeals          = "meals"
	CategoryOfficeSupplies = "office supplies"
)
