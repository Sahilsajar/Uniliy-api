package main

import (
	"context"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/unilly-api/api"
	"github.com/unilly-api/controllers"
	dbConn "github.com/unilly-api/db"
	db "github.com/unilly-api/db/sqlc"
	"github.com/unilly-api/repositories"
	"github.com/unilly-api/routes"
	"github.com/unilly-api/services"
)

func main() {
	_ = godotenv.Load() // Load .env file
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(api.Recovery())
	r.Use(api.ErrorHandler())
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
		api.Success(c, 200, "Unilly API is running", nil)
	})

	authRepo := repositories.NewAuthRepo(queries)
	authService := services.NewAuthService(authRepo)
	authController := controllers.NewAuthController(authService)

	routes.AuthRoutes(r, authController)

	r.Run("0.0.0.0:8080")
}
