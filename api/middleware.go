package api

import (
	"log"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		appErr := NormalizeError(c.Errors.Last().Err)
		c.JSON(appErr.StatusCode, errorResponse{
			Success: false,
			Error: responseError{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("panic recovered: %v", recovered)
		_ = c.Error(Internal("INTERNAL_SERVER_ERROR", "Something went wrong"))
		c.Abort()
	})
}
