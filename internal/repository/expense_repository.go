package repository

import (
	"flypro-assessment-ayo-rt/internal/models"

	"gorm.io/gorm"
)

type ExpenseRepository interface {
	Create(expense *models.Expense) error
	GetByID(id uint) (*models.Expense, error)
	GetByUserID(userID uint, offset, limit int, category, status string) ([]models.Expense, int64, error)
	Update(expense *models.Expense) error
	Delete(id uint) error
	GetByIDs(ids []uint) ([]models.Expense, error)
	GetByUserIDAndIDs(userID uint, ids []uint) ([]models.Expense, error)
}

type expenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) Create(expense *models.Expense) error {	return r.db.Create(expense).Error
}

func (r *expenseRepository) GetByID(id uint) (*models.Expense, error) {
	var expense models.Expense
	if err := r.db.Preload("User").First(&expense, id).Error; err != nil {
return nil, err
	}
	return &expense, nil
}
func (r *expenseRepository) GetByUserID(userID uint, offset, limit int, category, status string) ([]models.Expense, int64, error) {
	var expenses []models.Expense
var total int64
query := r.db.Model(&models.Expense{}).Where("user_id = ?", userID)

	if category != "" {
		query = query.Where("category = ?", category)
	}
if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&expenses).Error; err != nil {
return nil, 0, err
	}

return expenses, total, nil
}

func (r *expenseRepository) Update(expense *models.Expense) error {	return r.db.Save(expense).Error
}

func (r *expenseRepository) Delete(id uint) error {return r.db.Delete(&models.Expense{}, id).Error
}

func (r *expenseRepository) GetByIDs(ids []uint) ([]models.Expense, error) {
	var expenses []models.Expense
if err := r.db.Where("id IN ?", ids).Find(&expenses).Error; err != nil {		return nil, err
	}
	return expenses, nil
}

func (r *expenseRepository) GetByUserIDAndIDs(userID uint, ids []uint) ([]models.Expense, error) {var expenses []models.Expense
	if err := r.db.Where("user_id = ? AND id IN ?", userID, ids).Find(&expenses).Error; err != nil {
		return nil, err
	}
	return expenses, nil
}