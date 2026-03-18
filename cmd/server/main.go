package main

import (
	"context"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/unilly-api/controllers"
	db "github.com/unilly-api/db/sqlc"
	dbConn "github.com/unilly-api/db"
	"github.com/unilly-api/repositories"
	"github.com/unilly-api/routes"
	"github.com/unilly-api/services"
)

func main() {
	r := gin.Default()
	ctx := context.Background()

	database, err := dbConn.NewDB(ctx)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	fmt.Println("Database connection established:", database)

	queries := db.New(database)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	authRepo := repositories.NewAuthRepo(queries)
	authService := services.NewAuthService(authRepo)
	authController := controllers.NewAuthController(authService)

	routes.AuthRoutes(r, authController)

	r.Run()
}
