package repository

import (
	"flypro-assessment-ayo-rt/internal/models"

	"gorm.io/gorm"
)

type ExpenseReportRepository interface {
	Create(report *models.ExpenseReport) error
	GetByID(id uint) (*models.ExpenseReport, error)
	GetByUserID(userID uint, offset, limit int, status string) ([]models.ExpenseReport, int64, error)
	Update(report *models.ExpenseReport) error
	Delete(id uint) error
	AddExpenses(reportID uint, expenseIDs []uint) error
	RemoveExpenses(reportID uint, expenseIDs []uint) error
	UpdateTotal(reportID uint, total float64) error
}

type expenseReportRepository struct {
	db *gorm.DB
}

func NewExpenseReportRepository(db *gorm.DB) ExpenseReportRepository {
	return &expenseReportRepository{db: db}
}

func (r *expenseReportRepository) Create(report *models.ExpenseReport) error {
	return r.db.Create(report).Error
}

func (r *expenseReportRepository) GetByID(id uint) (*models.ExpenseReport, error) {
	var report models.ExpenseReport
	if err := r.db.Preload("User").Preload("Expenses").First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *expenseReportRepository) GetByUserID(userID uint, offset, limit int, status string) ([]models.ExpenseReport, int64, error) {
	var reports []models.ExpenseReport
	var total int64

	query := r.db.Model(&models.ExpenseReport{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Expenses").Offset(offset).Limit(limit).Order("created_at DESC").Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (r *expenseReportRepository) Update(report *models.ExpenseReport) error {
	return r.db.Save(report).Error
}

func (r *expenseReportRepository) Delete(id uint) error {
	return r.db.Delete(&models.ExpenseReport{}, id).Error
}

func (r *expenseReportRepository) AddExpenses(reportID uint, expenseIDs []uint) error {
	var expenses []models.Expense
	if err := r.db.Where("id IN ?", expenseIDs).Find(&expenses).Error; err != nil {
		return err
	}

	var report models.ExpenseReport
	if err := r.db.First(&report, reportID).Error; err != nil {
		return err
	}
	return r.db.Model(&report).Association("Expenses").Append(expenses)
}

func (r *expenseReportRepository) RemoveExpenses(reportID uint, expenseIDs []uint) error {
	var expenses []models.Expense
	if err := r.db.Where("id IN ?", expenseIDs).Find(&expenses).Error; err != nil {
		return err
	}
	return r.db.Model(&models.ExpenseReport{ID: reportID}).
		Association("Expenses").
		Delete(expenses)
}

func (r *expenseReportRepository) UpdateTotal(reportID uint, total float64) error {
	return r.db.Model(&models.ExpenseReport{}).
		Where("id = ?", reportID).
		Update("total", total).Error
}
