package api

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/unilly-api/utility"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		appErr := NormalizeError(c.Errors.Last().Err)
		c.JSON(appErr.StatusCode, errorResponse{
			StatusCode: appErr.StatusCode,
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

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(401, errorResponse{
				StatusCode: 401,
				Success: false,
				Error: responseError{
					Code:    "UNAUTHORIZED",
					Message: "Missing Authorization header",
				},
			})
			return
		}
		token := strings.Split(tokenStr, "Bearer ")[1]
		userID, err := utility.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, errorResponse{
				Success: false,
				Error: responseError{
					Code:    "INVALID_TOKEN",
					Message: "Invalid or expired token",
				},
			})
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}
