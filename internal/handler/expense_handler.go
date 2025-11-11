package handlers

import (
	"net/http"
	"strconv"

	"flypro-assessment-ayo-rt/internal/dto"
	"flypro-assessment-ayo-rt/internal/utils"
	"flypro-assessment-ayo-rt/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ExpenseHandler struct {
	expenseService services.ExpenseService
	validator      *validator.Validate
}

func NewExpenseHandler(expenseService services.ExpenseService, validator *validator.Validate) *ExpenseHandler {
	return &ExpenseHandler{
		expenseService: expenseService,
		validator:      validator,
	}
}

func (h *ExpenseHandler) CreateExpense(c *gin.Context) {
	var req dto.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		handleValidationError(c, err)
		return
	}

	userID := getUserIDFromContext(c)

	expense, err := h.expenseService.CreateExpense(c.Request.Context(), userID, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.ErrorResponse(c, appErr)
			return
		}
		utils.ErrorResponse(c, utils.NewInternalError("failed to create expense", err))
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, expense)
}

func (h *ExpenseHandler) GetExpense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, utils.NewBadRequestError("invalid expense ID"))
		return
	}

	expense, err := h.expenseService.GetExpenseByID(c.Request.Context(), uint(id))
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.ErrorResponse(c, appErr)
			return
		}
		utils.ErrorResponse(c, utils.NewInternalError("failed to get expense", err))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, expense)
}

func (h *ExpenseHandler) ListExpenses(c *gin.Context) {
	var filter dto.ExpenseFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		handleValidationError(c, err)
		return
	}

	userID := getUserIDFromContext(c)

	result, err := h.expenseService.ListExpenses(c.Request.Context(), userID, filter)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.ErrorResponse(c, appErr)
			return
		}
		utils.ErrorResponse(c, utils.NewInternalError("failed to list expenses", err))
		return
	}

	meta := utils.PaginationMeta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: int((result.Total + int64(result.PerPage) - 1) / int64(result.PerPage)),
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, result.Expenses, meta)
}

func (h *ExpenseHandler) UpdateExpense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, utils.NewBadRequestError("invalid expense ID"))
		return
	}

	var req dto.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		handleValidationError(c, err)
		return
	}

	userID := getUserIDFromContext(c)

	expense, err := h.expenseService.UpdateExpense(c.Request.Context(), uint(id), userID, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.ErrorResponse(c, appErr)
			return
		}
		utils.ErrorResponse(c, utils.NewInternalError("failed to update expense", err))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, expense)
}

func (h *ExpenseHandler) DeleteExpense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, utils.NewBadRequestError("invalid expense ID"))
		return
	}

	userID := getUserIDFromContext(c)

	if err := h.expenseService.DeleteExpense(c.Request.Context(), uint(id), userID); err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.ErrorResponse(c, appErr)
			return
		}
		utils.ErrorResponse(c, utils.NewInternalError("failed to delete expense", err))
		return
	}

	c.Status(http.StatusNoContent)
}