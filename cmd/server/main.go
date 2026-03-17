package main

import (
	"context"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/unilly-api/db"
	"github.com/unilly-api/routes"
)

func main() {
	r := gin.Default()
	ctx := context.Background()

	db, err := db.NewDB(ctx)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	fmt.Println("Database connection established:", db)

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
