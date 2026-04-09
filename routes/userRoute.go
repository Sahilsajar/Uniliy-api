package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/unilly-api/api"
	"github.com/unilly-api/controllers"
)

func UserRoutes(r *gin.Engine, userController *controllers.UserController) {
	userGroup := r.Group("/users")
	userGroup.Use(api.AuthMiddleware())
	userGroup.GET("/followers", api.Wrap(userController.GetFollowers))
	userGroup.GET("/following", api.Wrap(userController.GetFollowing))
}
