package handlers

import (
	"strconv"

	"flypro-assessment-ayo-rt/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func getUserIDFromContext(c *gin.Context) uint {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		return 1
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 1
	}

	return uint(userID)
}

func handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, fieldErr := range validationErrs {
			field := fieldErr.Field()
			tag := fieldErr.Tag()
			var message string

			switch tag {
			case "required":
				message = field + " is required"
			case "email":
				message = field + " must be a valid email"
			case "min":
				message = field + " must be at least " + fieldErr.Param() + " characters"
			case "max":
				message = field + " must be at most " + fieldErr.Param() + " characters"
			case "gt":
				message = field + " must be greater than " + fieldErr.Param()
			case "len":
				message = field + " must be exactly " + fieldErr.Param() + " characters"
			case "currency":
				message = field + " must be a valid currency code"
			case "oneof":
				message = field + " must be one of: " + fieldErr.Param()
			default:
				message = field + " is invalid"
			}

			errors[field] = message
		}
		utils.ValidationErrorResponse(c, errors)
	} else {
		utils.ErrorResponse(c, utils.NewBadRequestError(err.Error()))
	}
}


