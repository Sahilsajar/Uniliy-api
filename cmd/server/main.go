package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/unilly-api/db"
	"github.com/unilly-api/models"
	"github.com/gin-contrib/cors"
	"github.com/unilly-api/routes"
)

func main() {
    r := gin.Default()
	database, err := db.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	database.AutoMigrate(&models.User{})
	fmt.Println("Database connection established:", database)

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
	routes.AuthRoutes(r)
	r.Run()
}
