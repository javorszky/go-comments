package main

import (
	"fmt"
	"github.com/labstack/echo/middleware"
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
	e.Use(middleware.Gzip())

	files, err := getTemplates("public/js/*.js", "public/views/*.html")

	if nil != err {
		log.Fatalf("Failed parsing templates: %v", err)
	}

	e.Renderer = &Template{
		templates: template.Must(template.ParseFiles(files...)),
	}

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

// getTemplates variadic function that takes any number of single glob patterns
func getTemplates(paths ...string) (templates []string, err error) {
	for _, path := range paths {
		files, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("error reading templates from this path: %v. Message: %v", path, err)
		}
		templates = append(templates, files...)
	}

	return templates, nil
}
