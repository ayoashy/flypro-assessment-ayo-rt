package handler

import (
	"net/http"
	"strconv"

	"flypro-assessment-ayo-rt/internal/dto"
	"flypro-assessment-ayo-rt/internal/services"
	"flypro-assessment-ayo-rt/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ExpenseReportHandler struct {
	reportService services.ExpenseReportService
	validator     *validator.Validate
}

func NewExpenseReportHandler(reportService services.ExpenseReportService, validator *validator.Validate) *ExpenseReportHandler {
	return &ExpenseReportHandler{
		reportService: reportService,
		validator:     validator,
	}
}

func (h *ExpenseReportHandler) CreateReport(c *gin.Context) {
	var req dto.CreateExpenseReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		handleValidationError(c, err)
		return
	}

	userID := getUserIDFromContext(c)

	report, err := h.reportService.CreateReport(c.Request.Context(), userID, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.ErrorResponse(c, appErr)
			return
		}
		utils.ErrorResponse(c, utils.NewInternalError("failed to create report", err))
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, report)
}

func (h *ExpenseReportHandler) GetReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, utils.NewBadRequestError("invalid report ID"))
		return
	}

	report, err := h.reportService.GetReportByID(c.Request.Context(), uint(id))
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.ErrorResponse(c, appErr)
			return
		}
		utils.ErrorResponse(c, utils.NewInternalError("failed to get report", err))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, report)
}

func (h *ExpenseReportHandler) ListReports(c *gin.Context) {
	var filter dto.ReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		handleValidationError(c, err)
		return
	}

	userID := getUserIDFromContext(c)

	result, err := h.reportService.ListReports(c.Request.Context(), userID, filter)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.ErrorResponse(c, appErr)
			return
		}
		utils.ErrorResponse(c, utils.NewInternalError("failed to list reports", err))
		return
	}

	meta := utils.PaginationMeta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: int((result.Total + int64(result.PerPage) - 1) / int64(result.PerPage)),
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, result.Reports, meta)
}

func (h *ExpenseReportHandler) AddExpensesToReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, utils.NewBadRequestError("invalid report ID"))
		return
	}

	var req dto.AddExpensesToReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		handleValidationError(c, err)
		return
	}

	userID := getUserIDFromContext(c)

	if err := h.reportService.AddExpensesToReport(c.Request.Context(), uint(id), userID, req); err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.ErrorResponse(c, appErr)
			return
		}
		utils.ErrorResponse(c, utils.NewInternalError("failed to add expenses to report", err))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ExpenseReportHandler) SubmitReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, utils.NewBadRequestError("invalid report ID"))
		return
	}

	userID := getUserIDFromContext(c)

	if err := h.reportService.SubmitReport(c.Request.Context(), uint(id), userID); err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.ErrorResponse(c, appErr)
			return
		}
		utils.ErrorResponse(c, utils.NewInternalError("failed to submit report", err))
		return
	}

	c.Status(http.StatusNoContent)
}
