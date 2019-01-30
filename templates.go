package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"html/template"
	"io"
	"path/filepath"
)

// Template struct for working with templates and echo
type Template struct {
	templates *template.Template
}

// Render function: Overridden Render function for templates
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Get variadic function that takes any number of single glob patterns
func Get(paths ...string) (templates []string, err error) {
	for _, path := range paths {
		files, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("error reading templates from this path: %v. Message: %v", path, err)
		}
		templates = append(templates, files...)
	}

	return templates, nil
}

func GetTemplateFiles() ([]string, error) {
	files, err := Get("public/js/*.js", "public/views/partials/*.html", "public/views/*.html")

	if nil != err {
		return nil, fmt.Errorf("failed parsing templates: %v", err)
	}

	return files, nil

}

func SetRenderer(e *echo.Echo) {
	files, err := GetTemplateFiles()

	if nil != err {
		log.Fatalf("Setting the renderer failed: %v", err)
	}

	e.Renderer = &Template{
		templates: template.Must(template.ParseFiles(files...)),
	}
}
