package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/unilly-api/api"
	"github.com/unilly-api/controllers"
)

func AuthRoutes(r *gin.Engine, authController *controllers.AuthController) {
	authGroup := r.Group("/auth")
	authGroup.POST("/signup", api.Wrap(authController.SignUp))
	authGroup.POST("/generate-otp", api.Wrap(authController.GenerateAndSendOTP))
	authGroup.POST("/verify-otp", api.Wrap(authController.VerifyOTP))
	authGroup.POST("/login", api.Wrap(authController.Login))
	authGroup.GET("/me", api.AuthMiddleware(), api.Wrap(authController.GetProfile))
}
