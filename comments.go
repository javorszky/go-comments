package main

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"os"
)

func main() {
	fmt.Println("hello world")

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Print("Port not in env, setting it to 8090")
		port = "8090"
	}

	e.Logger.Fatal(e.Start(":" + port))
}