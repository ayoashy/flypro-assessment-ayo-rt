package services

import (
	"context"
	"fmt"
	"time"

	"flypro-assessment-ayo-rt/internal/config"
	"flypro-assessment-ayo-rt/internal/dto"
	"flypro-assessment-ayo-rt/internal/models"
	"flypro-assessment-ayo-rt/internal/repository"
	"flypro-assessment-ayo-rt/internal/utils"

	"github.com/redis/go-redis/v9"
)

type ExpenseService interface {
	CreateExpense(ctx context.Context, userID uint, req dto.CreateExpenseRequest) (*dto.ExpenseResponse, error)
	GetExpenseByID(ctx context.Context, id uint) (*dto.ExpenseResponse, error)
	ListExpenses(ctx context.Context, userID uint, filter dto.ExpenseFilter) (*dto.ExpenseListResponse, error)
	UpdateExpense(ctx context.Context, id, userID uint, req dto.UpdateExpenseRequest) (*dto.ExpenseResponse, error)
	DeleteExpense(ctx context.Context, id, userID uint) error
}

type expenseService struct {
	expenseRepo  repository.ExpenseRepository
	currencyService CurrencyService
	redisClient  *redis.Client
	config       *config.Config
}

func NewExpenseService(
	expenseRepo repository.ExpenseRepository,
	currencyService CurrencyService,
	redisClient *redis.Client,
	cfg *config.Config,
) ExpenseService {
	return &expenseService{
		expenseRepo:     expenseRepo,
		currencyService: currencyService,
		redisClient:     redisClient,
		config:          cfg,
	}
}

func (s *expenseService) CreateExpense(ctx context.Context, userID uint, req dto.CreateExpenseRequest) (*dto.ExpenseResponse, error) {
	expense := &models.Expense{
		UserID:      userID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Category:    req.Category,
		Description: req.Description,
		Receipt:     req.Receipt,
		Status:      models.ExpenseStatusPending,
	}

	if err := s.expenseRepo.Create(expense); err != nil {
		return nil, utils.NewInternalError("failed to create expense", err)
	}

	cacheKey := fmt.Sprintf("expenses:user:%d", userID)
	s.redisClient.Del(ctx, cacheKey)

	return s.mapToExpenseResponse(expense), nil
}

func (s *expenseService) GetExpenseByID(ctx context.Context, id uint) (*dto.ExpenseResponse, error) {
	expense, err := s.expenseRepo.GetByID(id)
	if err != nil {
		return nil, utils.NewNotFoundError("expense")
	}

	return s.mapToExpenseResponse(expense), nil
}

func (s *expenseService) ListExpenses(ctx context.Context, userID uint, filter dto.ExpenseFilter) (*dto.ExpenseListResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 10
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	offset := (filter.Page - 1) * filter.PerPage

	expenses, total, err := s.expenseRepo.GetByUserID(userID, offset, filter.PerPage, filter.Category, filter.Status)
	if err != nil {
		return nil, utils.NewInternalError("failed to list expenses", err)
	}

	expenseResponses := make([]dto.ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		expenseResponses[i] = *s.mapToExpenseResponse(&expense)
	}

	return &dto.ExpenseListResponse{
		Expenses: expenseResponses,
		Page:     filter.Page,
		PerPage:  filter.PerPage,
		Total:    total,
	}, nil
}

func (s *expenseService) UpdateExpense(ctx context.Context, id, userID uint, req dto.UpdateExpenseRequest) (*dto.ExpenseResponse, error) {
	expense, err := s.expenseRepo.GetByID(id)
	if err != nil {
		return nil, utils.NewNotFoundError("expense")
	}

	if expense.UserID != userID {
		return nil, utils.NewBadRequestError("you can only update your own expenses")
	}
	if req.Amount != nil {
		expense.Amount = *req.Amount
	}
	if req.Currency != "" {
		expense.Currency = req.Currency
	}
	if req.Category != "" {
		expense.Category = req.Category
	}
	if req.Description != "" {
		expense.Description = req.Description
	}
	if req.Receipt != "" {
		expense.Receipt = req.Receipt
	}
	if req.Status != "" {
		expense.Status = req.Status
	}

	if err := s.expenseRepo.Update(expense); err != nil {
		return nil, utils.NewInternalError("failed to update expense", err)
	}

	cacheKey := fmt.Sprintf("expenses:user:%d", userID)
	s.redisClient.Del(ctx, cacheKey)

	return s.mapToExpenseResponse(expense), nil
}

func (s *expenseService) DeleteExpense(ctx context.Context, id, userID uint) error {
	expense, err := s.expenseRepo.GetByID(id)
	if err != nil {
		return utils.NewNotFoundError("expense")
	}

	if expense.UserID != userID {
		return utils.NewBadRequestError("you can only delete your own expenses")
	}

	if err := s.expenseRepo.Delete(id); err != nil {
		return utils.NewInternalError("failed to delete expense", err)
	}

	cacheKey := fmt.Sprintf("expenses:user:%d", userID)
	s.redisClient.Del(ctx, cacheKey)

	return nil
}

func (s *expenseService) mapToExpenseResponse(expense *models.Expense) *dto.ExpenseResponse {
	return &dto.ExpenseResponse{
		ID:          expense.ID,
		UserID:      expense.UserID,
		Amount:      expense.Amount,
		Currency:    expense.Currency,
		Category:    expense.Category,
		Description: expense.Description,
		Receipt:     expense.Receipt,
		Status:      expense.Status,
		CreatedAt:   expense.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   expense.UpdatedAt.Format(time.RFC3339),
	}
}