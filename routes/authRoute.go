package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {
	r.GET("/auth", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"Success": "true", "msg": "go away lad"})
	})
}
