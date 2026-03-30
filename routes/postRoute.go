package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/unilly-api/api"
	"github.com/unilly-api/controllers"
)

func PostRoutes(r *gin.Engine, postController *controllers.PostController) {
	postGroup := r.Group("/posts")
	postGroup.Use(api.AuthMiddleware())
	postGroup.POST("/media/upload", api.Wrap(postController.UploadTempMedia))
	postGroup.POST("", api.Wrap(postController.CreatePost))
	// postGroup.POST("/:postID/tags", api.Wrap(postController.TagUsers))
}
