package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	mocket "github.com/selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	serveJS            = `console.log("The id of the thing requesting this was 44");`
	mockGoodUser       = `{"email":"test@example.com","name":"John Doe","password1":"somepassword", "password2":"somepassword", "csrf":"somevalue"}`
	mockGoodUserReturn = `{"ID":0,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","DeletedAt":null,"email":"test@example.com","name":"John Doe","password1":"somepassword","password2":"somepassword"}`

	mockBadUser       = `{"email":"test@example.com","name":"John Doe","password1":"somepassword", "password2":"someotherpass"}`
	mockBadUserReturn = `{"error":"Passwords do not match."}`

	mockNetworkErrorUser       = `{"email":"test@example.com","name":"John Doe","password1":"NetworkError", "password2":"NetworkError"}`
	mockNetworkErrorUserReturn = `{"error":"HTTP request failed with error: Unavailable"}`

	mockFoundPassUser       = `{"email":"test@example.com","name":"John Doe","password1":"FoundPassword", "password2":"FoundPassword"}`
	mockFoundPassUserReturn = `{"error":"Password is found in the database."}`

	e    *echo.Echo
	mpwc MockPasswordChecker
	pwh  PasswordHasher
	db   *gorm.DB
	h    Handlers
)

type MockPasswordChecker struct{}
type MockPasswordHasher struct{}

func (mpwc MockPasswordChecker) IsPasswordPwnd(password string) (bool, error) {
	switch password {
	case "NetworkError":
		return false, fmt.Errorf("HTTP request failed with error: %s", "Unavailable")
	case "FoundPassword":
		return true, nil
	default:
		return false, nil
	}
}

func _CSRFSkipper(echo.Context) bool {
	return true
}

func (mpwh MockPasswordHasher) GenerateFromPassword(password string) (string, error) {
	if password == "canthashthis" {
		return "", errors.New("hashing password failed")
	}

	return "hashedpassword", nil
}

func (mpwh MockPasswordHasher) ComparePasswordAndHash(password string, hash string) (bool, error) {
	if "cantcomparethis" == password {
		return false, errors.New("failed comparing password and hash")
	} else if "goodpassword" == password {
		return true, nil
	} else {
		return false, nil
	}
}

func TestMain(m *testing.M) {
	e = echo.New()
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		Skipper: _CSRFSkipper,
	}))

	mpwc = MockPasswordChecker{}

	pwh = MockPasswordHasher{}

	mocket.Catcher.Register() // Safe register. Allowed multiple calls to save
	mocket.Catcher.Logging = true
	// GORM
	DB, err := gorm.Open(mocket.DriverName, "connection_string") // Can be any connection string
	if err != nil {
		log.Fatal("Mocket failed to initialise.")
	}

	db = DB

	h = NewHandler(mpwc, pwh, db)
	SetRenderer(e)

	os.Exit(m.Run())
}

func TestServeJS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/44/js", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/:id/js")
	c.SetParamNames("id")
	c.SetParamValues("44")

	if assert.NoError(t, h.ServeJS(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, serveJS, rec.Body.String())
	}
}

func TestPageRenders(t *testing.T) {
	pairs := []struct {
		Path         string
		ExpectedCode int
	}{
		{"/", http.StatusOK},
		{"/login", http.StatusOK},
		{"/register", http.StatusOK},
	}

	for _, r := range pairs {
		req := httptest.NewRequest(http.MethodGet, r.Path, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetPath(r.Path)

		if assert.NoError(t, h.Index(c)) {
			assert.Equal(t, r.ExpectedCode, rec.Code)
		}
	}
}

func TestRegisterPostGood(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(mockGoodUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register")

	if assert.NoError(t, h.RegisterPost(c)) {
		var origin map[string]interface{}
		if err := json.Unmarshal([]byte(mockGoodUser), &origin); err != nil {
			panic(err)
		}

		var dat map[string]interface{}

		if err := json.Unmarshal(rec.Body.Bytes(), &dat); err != nil {
			panic(err)
		}

		timeCreatedAt, err := time.Parse(time.RFC3339Nano, dat["CreatedAt"].(string))
		if err != nil {
			assert.Fail(t, "could not parse time")
		}

		fmt.Println(timeCreatedAt)

		timeNow := time.Now()
		timeString := fmt.Sprintf("%4d-%02d-%02dT%02d:%02d:%02d", timeNow.Year(), timeNow.Month(), timeNow.Day(), timeNow.Hour(), timeNow.Minute(), timeNow.Second())

		created := strings.HasPrefix(dat["CreatedAt"].(string), timeString)
		updated := strings.HasPrefix(dat["UpdatedAt"].(string), timeString)

		assert.True(t, created)
		assert.True(t, updated)
		assert.NotNil(t, dat["ID"])
		assert.Nil(t, dat["DeletedAt"])
		assert.Equal(t, origin["CreatedAt"], origin["UpdatedAt"])
		assert.Equal(t, origin["email"], dat["email"])
		assert.Equal(t, "hashedpassword", dat["passwordHash"])
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestRegisterPostPasswordDontMatch(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(mockBadUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register")

	if assert.NoError(t, h.RegisterPost(c)) {
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Equal(t, mockBadUserReturn, rec.Body.String())
	}
}

func TestRegisterPostNetworkError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(mockNetworkErrorUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register")

	if assert.NoError(t, h.RegisterPost(c)) {
		assert.Equal(t, http.StatusBadGateway, rec.Code)
		assert.Equal(t, mockNetworkErrorUserReturn, rec.Body.String())
	}
}

func TestRegisterPasswordPawned(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(mockFoundPassUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register")

	if assert.NoError(t, h.RegisterPost(c)) {
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Equal(t, mockFoundPassUserReturn, rec.Body.String())
	}
}
