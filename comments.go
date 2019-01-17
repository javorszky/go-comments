package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo"
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
	e := echo.New()

	var templates []string

	js, err := filepath.Glob("public/js/*.js")
	if err != nil {
		log.Fatal(err)
	}

	html, err := filepath.Glob("public/views/*.html")
	if err != nil {
		log.Fatal(err)
	}

	templates = append(templates, js...)
	templates = append(templates, html...)

	t := &Template{
		templates: template.Must(template.ParseFiles(templates...)),
	}

	e.Renderer = t

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", "")
	})

	e.GET("/:id/js", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJavaScript)
		return c.Render(http.StatusOK, "client.js", c.Param("id"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Print("Port not in env, setting it to 8090")
		port = "8090"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
