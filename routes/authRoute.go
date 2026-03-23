package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/unilly-api/controllers"
)

func AuthRoutes(r *gin.Engine, authController *controllers.AuthController) {
	authGroup := r.Group("/auth")
	// authGroup.POST("/signup", authController.SignUp)
	authGroup.POST("/generate-otp", authController.GenerateAndSendOTP)
	authGroup.POST("/verify-otp", authController.VerifyOTP)
}
