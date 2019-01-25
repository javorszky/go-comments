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
	"html/template"
	"io"
	"log"
)

// Template struct for working with templates and echo
type Template struct {
	templates *template.Template
}

// Render function: Overridden Render function for templates
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

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

	files, err := templates.Get("public/js/*.js", "public/views/partials/*.html", "public/views/*.html")

	if nil != err {
		log.Fatalf("Failed parsing templates: %v", err)
	}

	e.Renderer = &Template{
		templates: template.Must(template.ParseFiles(files...)),
	}

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
