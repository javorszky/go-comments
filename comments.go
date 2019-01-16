package main

import (
	"fmt"
	"github.com/labstack/echo"
	"os"
)

func main() {
	fmt.Println("hello world")

	e := echo.New()

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Print("Port not in env, setting it to 8090")
		port = "8090"
	}

	e.Logger.Fatal(e.Start(":" + port))
}