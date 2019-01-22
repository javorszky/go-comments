package main

import (
	"fmt"
	"github.com/javorszky/go-comments/config"
	"github.com/javorszky/go-comments/templates"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo/middleware"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo"
)

// Template struct for working with templates and echo
type Template struct {
	templates *template.Template
}

type User struct {
	gorm.Model
	Email        string
	PasswordHash string
}

// Render function: Overridden Render function for templates
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	// Config
	config, err := config.Get()

	if err != nil {
		log.Fatalf("Failed getting config. Error was %v", err)
	}

	// Database
	db, err := gorm.Open("mysql", fmt.Sprintf("%v:%v@/%v?charset=utf8mb4&parseTime=True&loc=Local", config.DatabaseUser, config.DatabasePassword, config.DatabaseTable))

	if err != nil {
		log.Fatalf("Failed connecting to database: %v", err)
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

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", "")
	})

	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login", "")
	})

	e.GET("/register", func(c echo.Context) error {
		return c.Render(http.StatusOK, "register", "")
	})

	e.GET("/:id/js", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJavaScript)
		return c.Render(http.StatusOK, "client.js", c.Param("id"))
	})

	e.GET("/request", func(c echo.Context) error {
		req := c.Request()
		format := `
<code>
Protocol: %s<br>
Host: %s<br>
Remote Address: %s<br>
Method: %s<br>
Path: %s<br>
TLS: %v<br>
TLS Version: %v<br>
</code>
`
		return c.HTML(http.StatusOK, fmt.Sprintf(format, req.Proto, req.Host, req.RemoteAddr, req.Method, req.URL.Path, req.TLS.NegotiatedProtocol, req.TLS.Version))
	})

	port := config.Port
	if port == "" {
		fmt.Print("Port not in env, setting it to 8090")
		port = "8090"
	}

	// e.Logger.Fatal(e.Start(":" + port))
	e.Logger.Fatal(e.StartTLS(":1323", "cert.crt", "key.key"))
}
