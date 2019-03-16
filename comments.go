package main

import (
	"fmt"
	"github.com/javorszky/go-comments/config"
	"github.com/javorszky/go-comments/db"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/go-playground/validator.v9"
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

	m := database.RunMigrations(db)

	if err = m; err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
	log.Printf("Migration did run successfully")

	e := echo.New()
	e.Use(middleware.Gzip())
	e.Validator = &CustomValidator{validator: validator.New()}
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

	pwhParams := argon2Params{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
	pwc := PwChecker{}
	pwh := NewArgon2(pwhParams)
	h := NewHandler(pwc, pwh, db)

	e.GET("/", h.Index)

	e.GET("/login", h.Login)

	e.POST("/login", h.LoginPost)

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
