package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

// func (ac *AuthController) SignUp(ctx *gin.Context) {
// 	user := &models.User{
// 		Username: "sahil",
// 		Email:    "sahil@example.com",
// 		PasswordHash: "Password123",
// 	}
// 	ac.authService.SignUp(ctx.Request.Context(), user)
// 	ctx.JSON(http.StatusOK, gin.H{"Success": "sign up successful"})
// }



func (ac *AuthController) GenerateAndSendOTP(ctx *gin.Context) {
	type VerifyEmailRequest struct {
		Email string `json:"email" binding:"required,email"`
	}
	var req VerifyEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := ac.authService.GenerateAndSendOTP(ctx.Request.Context(), req.Email); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})

}

func (ac *AuthController) VerifyOTP(ctx *gin.Context) {
	type VerifyOTPRequest struct {
		Email string `json:"email" binding:"required,email"`
		OTP   string `json:"otp" binding:"required,len=6"`
	}
	var req VerifyOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	valid, err := ac.authService.VerifyOTP(ctx.Request.Context(), req.Email, req.OTP)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify OTP"})
		return
	}
	if !valid {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})

}

func (ac *AuthController) Login(ctx *gin.Context) {
	type LoginRequest struct {
		Email string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	accessToken, refreshToken,err := ac.authService.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Login successful", "access_token": accessToken, "refresh_token": refreshToken})	
}