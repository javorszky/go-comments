package main

import (
	"fmt"
	"github.com/javorszky/go-comments/config"
	"github.com/javorszky/go-comments/db"
	"github.com/javorszky/go-comments/handlers"
	"github.com/javorszky/go-comments/templates"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"log"
)

func main() {
	// Config
	config, err := config.Get()

	if err != nil {
		log.Fatalf("Failed getting config: %v", err)
	}

	db, err := database.Get(config)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer db.Close()

	e := echo.New()
	e.Use(middleware.Gzip())
	e.Static("/static", "public/static")

	templates.SetRenderer(e)

	e.GET("/", handlers.Index)

	e.GET("/login", handlers.Login)

	e.POST("/register", handlers.RegisterPost)

	e.GET("/register", handlers.Register)

	e.GET("/:id/js", handlers.ServeJS)

	e.GET("/request", handlers.Request)

	port := config.Port
	if port == "" {
		fmt.Print("Port not in env, setting it to 8090")
		port = "8090"
	}

	// e.Logger.Fatal(e.Start(":" + port))
	e.Logger.Fatal(e.StartTLS(":1323", "cert.crt", "key.key"))
}
