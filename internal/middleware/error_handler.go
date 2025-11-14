package middleware

import (
	"net/http"

	"flypro-assessment-ayo-rt/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if appErr, ok := err.(*utils.AppError); ok {
				utils.ErrorResponse(c, appErr)
				return
			}

			logger.Error("Unhandled error",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)

			utils.ErrorResponse(c, utils.NewInternalError("internal server error", err))
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}
