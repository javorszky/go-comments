package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

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

	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.js")),
	}

	e.Renderer = t

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/:id/js", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJavaScript)
		return c.Render(http.StatusOK, "clientjs", c.Param("id"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Print("Port not in env, setting it to 8090")
		port = "8090"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
