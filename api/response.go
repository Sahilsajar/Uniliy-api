package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type successResponse struct {
	StatusCode  int    `json:"statuscode"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type errorResponse struct {
	StatusCode  int           `json:"statuscode"`
	Success bool          `json:"success"`
	Error   responseError `json:"error"`
}

type responseError struct {
	StatusCode  int    `json:"statuscode"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type validationField struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

type HandlerFunc func(*gin.Context) error

func Wrap(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := handler(c); err != nil {
			_ = c.Error(NormalizeError(err))
			c.Abort()
		}
	}
}

func Success(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, successResponse{
		StatusCode: statusCode,
		Success: true,
		Message: message,
		Data:    data,
	})
}

func BindJSON(c *gin.Context, target any) error {
	if err := c.ShouldBindJSON(target); err != nil {
		return BadRequest("VALIDATION_ERROR", "Invalid request body").WithDetails(validationDetails(err)).WithCause(err)
	}

	return nil
}

func validationDetails(err error) any {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		fields := make([]validationField, 0, len(validationErrs))
		for _, validationErr := range validationErrs {
			fields = append(fields, validationField{
				Field: validationErr.Field(),
				Rule:  validationErr.Tag(),
			})
		}
		return fields
	}

	return err.Error()
}
