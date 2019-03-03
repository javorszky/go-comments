package main

import (
	"fmt"
	"github.com/javorszky/go-comments/config"
	"github.com/javorszky/go-comments/db"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"log"
)

func main() {
	// Config
	localConfig, err := config.Get()

	if err != nil {
		log.Fatalf("Failed getting config: %v", err)
	}

	db, err := database.GetInstance(localConfig)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer db.Close()

	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:  "form:csrf",
		TokenLength:  128,
		CookieName:   "_csrf",
		CookieMaxAge: 300,
		CookieSecure: true,
		CookiePath:   "/",
	}))

	e.Static("/static", "public/static")

	SetRenderer(e)

	pwc := PwChecker{}
	pwh := Argon2{}
	pwhParams := argon2Params{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
	pwh.Init(pwhParams)

	h := NewHandler(pwc, &pwh)

	e.GET("/", h.Index)

	e.GET("/login", h.Login)

	e.POST("/register", h.RegisterPost)

	e.GET("/register", h.Register)

	e.GET("/:id/js", h.ServeJS)

	e.GET("/request", h.Request)
	port := localConfig.Port
	if port == "" {
		fmt.Print("Port not in env, setting it to 8090")
		port = "8090"
	}

	// e.Logger.Fatal(e.Start(":" + port))
	e.Logger.Fatal(e.StartTLS(":1323", "cert.crt", "key.key"))
}
