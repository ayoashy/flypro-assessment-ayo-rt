package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"flypro-assessment-ayo-rt/internal/config"
	"flypro-assessment-ayo-rt/internal/dto"
	"flypro-assessment-ayo-rt/internal/models"
	"flypro-assessment-ayo-rt/internal/repository"
	"flypro-assessment-ayo-rt/internal/utils"

	"github.com/redis/go-redis/v9"
)

type ExpenseReportService interface {
	CreateReport(ctx context.Context, userID uint, req dto.CreateExpenseReportRequest) (*dto.ExpenseReportResponse, error)
	GetReportByID(ctx context.Context, id uint) (*dto.ExpenseReportResponse, error)
	ListReports(ctx context.Context, userID uint, filter dto.ReportFilter) (*dto.ExpenseReportListResponse, error)
	AddExpensesToReport(ctx context.Context, reportID, userID uint, req dto.AddExpensesToReportRequest) error
	SubmitReport(ctx context.Context, reportID, userID uint) error
}

type expenseReportService struct {
	reportRepo      repository.ExpenseReportRepository
	expenseRepo     repository.ExpenseRepository
	currencyService CurrencyService
	redisClient     *redis.Client
	config          *config.Config
}

func NewExpenseReportService(
	reportRepo repository.ExpenseReportRepository,
	expenseRepo repository.ExpenseRepository,
	currencyService CurrencyService,
	redisClient *redis.Client,
	cfg *config.Config,
) ExpenseReportService {
	return &expenseReportService{
		reportRepo:      reportRepo,
		expenseRepo:     expenseRepo,
		currencyService: currencyService,
		redisClient:     redisClient,
		config:          cfg,
	}
}

func (s *expenseReportService) CreateReport(ctx context.Context, userID uint, req dto.CreateExpenseReportRequest) (*dto.ExpenseReportResponse, error) {
	report := &models.ExpenseReport{
		UserID: userID,
		Title:  req.Title,
		Status: models.ReportStatusDraft,
		Total:  0,
	}

	if err := s.reportRepo.Create(report); err != nil {
		return nil, utils.NewInternalError("failed to create expense report", err)
	}

	cacheKey := fmt.Sprintf("reports:user:%d", userID)
	s.redisClient.Del(ctx, cacheKey)

	return s.mapToReportResponse(report), nil
}

func (s *expenseReportService) GetReportByID(ctx context.Context, id uint) (*dto.ExpenseReportResponse, error) {
	cacheKey := fmt.Sprintf("report:%d", id)
	cachedReport, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var report models.ExpenseReport
		if err := json.Unmarshal([]byte(cachedReport), &report); err == nil {
			return s.mapToReportResponse(&report), nil
		}
	}

	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		return nil, utils.NewNotFoundError("expense report")
	}

	reportJSON, _ := json.Marshal(report)
	s.redisClient.Set(ctx, cacheKey, reportJSON, 30*time.Minute)

	return s.mapToReportResponse(report), nil
}

func (s *expenseReportService) ListReports(ctx context.Context, userID uint, filter dto.ReportFilter) (*dto.ExpenseReportListResponse, error) {
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

	reports, total, err := s.reportRepo.GetByUserID(userID, offset, filter.PerPage, filter.Status)
	if err != nil {
		return nil, utils.NewInternalError("failed to list reports", err)
	}

	reportResponses := make([]dto.ExpenseReportResponse, len(reports))
	for i, report := range reports {
		reportResponses[i] = *s.mapToReportResponse(&report)
	}

	return &dto.ExpenseReportListResponse{
		Reports: reportResponses,
		Page:    filter.Page,
		PerPage: filter.PerPage,
		Total:   total,
	}, nil
}

func (s *expenseReportService) AddExpensesToReport(ctx context.Context, reportID, userID uint, req dto.AddExpensesToReportRequest) error {
	report, err := s.reportRepo.GetByID(reportID)
	if err != nil {
		return utils.NewNotFoundError("expense report")
	}

	if report.UserID != userID {
		return utils.NewBadRequestError("you can only modify your own reports")
	}

	if report.Status != models.ReportStatusDraft {
		return utils.NewBadRequestError("can only add expenses to draft reports")
	}

	expenses, err := s.expenseRepo.GetByUserIDAndIDs(userID, req.ExpenseIDs)
	if err != nil {
		return utils.NewInternalError("failed to fetch expenses", err)
	}

	if len(expenses) != len(req.ExpenseIDs) {
		return utils.NewBadRequestError("some expenses not found or do not belong to you")
	}

	if err := s.reportRepo.AddExpenses(reportID, req.ExpenseIDs); err != nil {
		return utils.NewInternalError("failed to add expenses to report", err)
	}

	total, err := s.calculateTotal(ctx, reportID)
	if err != nil {
		return utils.NewInternalError("failed to calculate total", err)
	}

	if err := s.reportRepo.UpdateTotal(reportID, total); err != nil {
		return utils.NewInternalError("failed to update total", err)
	}

	s.redisClient.Del(ctx, fmt.Sprintf("report:%d", reportID))
	s.redisClient.Del(ctx, fmt.Sprintf("reports:user:%d", userID))

	return nil
}

func (s *expenseReportService) SubmitReport(ctx context.Context, reportID, userID uint) error {
	report, err := s.reportRepo.GetByID(reportID)
	if err != nil {
		return utils.NewNotFoundError("expense report")
	}

	if report.UserID != userID {
		return utils.NewBadRequestError("you can only submit your own reports")
	}

	if report.Status != models.ReportStatusDraft {
		return utils.NewBadRequestError("can only submit draft reports")
	}

	if len(report.Expenses) == 0 {
		return utils.NewBadRequestError("cannot submit report without expenses")
	}

	report.Status = models.ReportStatusSubmitted
	if err := s.reportRepo.Update(report); err != nil {
		return utils.NewInternalError("failed to submit report", err)
	}

	s.redisClient.Del(ctx, fmt.Sprintf("report:%d", reportID))
	s.redisClient.Del(ctx, fmt.Sprintf("reports:user:%d", userID))

	return nil
}

func (s *expenseReportService) calculateTotal(ctx context.Context, reportID uint) (float64, error) {
	report, err := s.reportRepo.GetByID(reportID)
	if err != nil {
		return 0, err
	}

	total := 0.0
	for _, expense := range report.Expenses {
		amountUSD, err := s.currencyService.ConvertCurrency(ctx, expense.Amount, expense.Currency, "USD")
		if err != nil {
			total += expense.Amount
		} else {
			total += amountUSD
		}
	}

	return total, nil
}

func (s *expenseReportService) mapToReportResponse(report *models.ExpenseReport) *dto.ExpenseReportResponse {
	expenseResponses := make([]dto.ExpenseResponse, len(report.Expenses))
	for i, expense := range report.Expenses {
		expenseResponses[i] = dto.ExpenseResponse{
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

	return &dto.ExpenseReportResponse{
		ID:        report.ID,
		UserID:    report.UserID,
		Title:     report.Title,
		Status:    report.Status,
		Total:     report.Total,
		CreatedAt: report.CreatedAt.Format(time.RFC3339),
		UpdatedAt: report.UpdatedAt.Format(time.RFC3339),
		Expenses:  expenseResponses,
	}
}
