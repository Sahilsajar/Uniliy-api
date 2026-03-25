package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unilly-api/api"
	"github.com/unilly-api/dto"
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

func (ac *AuthController) SignUp(ctx *gin.Context) error {
	var user dto.CreateUserRequestDTO
	if err := api.BindJSON(ctx, &user); err != nil {
		return err
	}
	if err := ac.authService.SignUp(ctx.Request.Context(), user); err != nil {
		return err
	}
	api.Success(ctx, http.StatusCreated, "Signup successful", nil)
	return nil
}

func (ac *AuthController) GenerateAndSendOTP(ctx *gin.Context) error {
	type VerifyEmailRequest struct {
		Email string `json:"email" binding:"required,email"`
	}
	var req VerifyEmailRequest
	if err := api.BindJSON(ctx, &req); err != nil {
		return err
	}
	if err := ac.authService.GenerateAndSendOTP(ctx.Request.Context(), req.Email); err != nil {
		return err
	}
	api.Success(ctx, http.StatusOK, "OTP sent successfully", nil)
	return nil
}

func (ac *AuthController) VerifyOTP(ctx *gin.Context) error {
	type VerifyOTPRequest struct {
		Email string `json:"email" binding:"required,email"`
		OTP   string `json:"otp" binding:"required,len=6"`
	}
	var req VerifyOTPRequest
	if err := api.BindJSON(ctx, &req); err != nil {
		return err
	}
	if err := ac.authService.VerifyOTP(ctx.Request.Context(), req.Email, req.OTP); err != nil {
		return err
	}
	api.Success(ctx, http.StatusOK, "OTP verified successfully", nil)
	return nil
}

func (ac *AuthController) Login(ctx *gin.Context) error {
	type LoginRequest struct {
		Identifier string `json:"identifier" binding:"required"`
		Password   string `json:"password" binding:"required"`
	}
	var req LoginRequest
	if err := api.BindJSON(ctx, &req); err != nil {
		return err
	}
	accessToken, refreshToken, err := ac.authService.Login(ctx.Request.Context(), req.Identifier, req.Password)
	if err != nil {
		return err
	}
	api.Success(ctx, http.StatusOK, "Login successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
	return nil
}

func (ac *AuthController) GetProfile(ctx *gin.Context) error {
	userID := ctx.GetString("user_id")
	userIDInt64, err := strconv.Atoi(userID)
	if err != nil {
		return api.BadRequest("INVALID_USER_ID", "Invalid user ID")
	}
	profile, err := ac.authService.GetProfile(ctx.Request.Context(), int64(userIDInt64))
	if err != nil {
		return err
	}
	api.Success(ctx, http.StatusOK, "Profile retrieved successfully", profile)
	return nil
}
func (ac *AuthController) RefreshToken(ctx *gin.Context) error {
	type RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	var req RefreshTokenRequest
	if err := api.BindJSON(ctx, &req); err != nil {
		return err
	}
	accessToken, refreshToken, err := ac.authService.RefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		return err
	}
	api.Success(ctx, http.StatusOK, "Token refreshed successfully", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
	return nil
}
