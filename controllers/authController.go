package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unilly-api/models"
	"github.com/unilly-api/services"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (ac *AuthController) SignUp(ctx *gin.Context) {
	user := &models.User{
		Username: "sahil",
		Email:    "sahil@example.com",
		PasswordHash: "Password123",
	}
	ac.authService.SignUp(ctx.Request.Context(), user)
	ctx.JSON(http.StatusOK, gin.H{"Success": "sign up successful"})
}
