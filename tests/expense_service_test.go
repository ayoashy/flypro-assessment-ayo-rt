package tests

import (
	"context"
	"errors"
	"testing"

	"flypro-assessment-ayo-rt/internal/dto"
	"flypro-assessment-ayo-rt/internal/models"
	"flypro-assessment-ayo-rt/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockExpenseRepository struct {
	mock.Mock
}

func (m *MockExpenseRepository) Create(expense *models.Expense) error {
	args := m.Called(expense)
	return args.Error(0)
}

func (m *MockExpenseRepository) GetByID(id uint) (*models.Expense, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetByUserID(userID uint, offset, limit int, category, status string) ([]models.Expense, int64, error) {
	args := m.Called(userID, offset, limit, category, status)
	return args.Get(0).([]models.Expense), args.Get(1).(int64), args.Error(2)
}

func (m *MockExpenseRepository) Update(expense *models.Expense) error {
	args := m.Called(expense)
	return args.Error(0)
}

func (m *MockExpenseRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockExpenseRepository) GetByIDs(ids []uint) ([]models.Expense, error) {
	args := m.Called(ids)
	return args.Get(0).([]models.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetByUserIDAndIDs(userID uint, ids []uint) ([]models.Expense, error) {
	args := m.Called(userID, ids)
	return args.Get(0).([]models.Expense), args.Error(1)
}

type MockCurrencyService struct {
	mock.Mock
}

func (m *MockCurrencyService) ConvertCurrency(ctx context.Context, amount float64, from, to string) (float64, error) {
	args := m.Called(ctx, amount, from, to)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCurrencyService) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(float64), args.Error(1)
}

func TestExpenseService_CreateExpense(t *testing.T) {
	tests := []struct {
		name          string
		input         dto.CreateExpenseRequest
		setupMocks    func(*MockExpenseRepository, *MockCurrencyService)
		expectedError error
		expectedID    uint
	}{
		{
			name: "valid expense creation",
			input: dto.CreateExpenseRequest{
				Amount:   100.50,
				Currency: "USD",
				Category: "travel",
			},
			setupMocks: func(mockRepo *MockExpenseRepository, mockCurrency *MockCurrencyService) {
				mockRepo.On("Create", mock.AnythingOfType("*models.Expense")).Return(nil).Run(func(args mock.Arguments) {
					expense := args.Get(0).(*models.Expense)
					expense.ID = 1
				})
			},
			expectedID: 1,
		},
		{
			name: "repository error",
			input: dto.CreateExpenseRequest{
				Amount:   100.50,
				Currency: "USD",
				Category: "travel",
			},
			setupMocks: func(mockRepo *MockExpenseRepository, mockCurrency *MockCurrencyService) {
				mockRepo.On("Create", mock.AnythingOfType("*models.Expense")).Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockExpenseRepository)
			mockCurrency := new(MockCurrencyService)

			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo, mockCurrency)
			}

			service := services.NewExpenseService(mockRepo, mockCurrency, nil, nil)
			result, err := service.CreateExpense(context.Background(), 1, tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedID, result.ID)
			}

			mockRepo.AssertExpectations(t)
			mockCurrency.AssertExpectations(t)
		})
	}
}

func TestExpenseService_GetExpenseByID(t *testing.T) {
	tests := []struct {
		name          string
		expenseID     uint
		setupMocks    func(*MockExpenseRepository)
		expectedError bool
	}{
		{
			name:      "expense found",
			expenseID: 1,
			setupMocks: func(mockRepo *MockExpenseRepository) {
				mockRepo.On("GetByID", uint(1)).Return(&models.Expense{
					ID:       1,
					UserID:   1,
					Amount:   100.50,
					Currency: "USD",
					Category: "travel",
					Status:   models.ExpenseStatusPending,
				}, nil)
			},
			expectedError: false,
		},
		{
			name:      "expense not found",
			expenseID: 999,
			setupMocks: func(mockRepo *MockExpenseRepository) {
				mockRepo.On("GetByID", uint(999)).Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockExpenseRepository)
			mockCurrency := new(MockCurrencyService)

			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			service := services.NewExpenseService(mockRepo, mockCurrency, nil, nil)
			result, err := service.GetExpenseByID(context.Background(), tt.expenseID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expenseID, result.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
