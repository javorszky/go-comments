package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var (
	serveJS            = `console.log("The id of the thing requesting this was 44");`
	mockGoodUser       = `{"email":"test@example.com","name":"John Doe","password1":"somepassword", "password2":"somepassword"}`
	mockGoodUserReturn = `{"ID":0,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","DeletedAt":null,"email":"test@example.com","name":"John Doe","password1":"somepassword","password2":"somepassword"}`

	mockBadUser       = `{"email":"test@example.com","name":"John Doe","password1":"somepassword", "password2":"someotherpass"}`
	mockBadUserReturn = `{"error":"Passwords do not match."}`

	mockNetworkErrorUser       = `{"email":"test@example.com","name":"John Doe","password1":"NetworkError", "password2":"NetworkError"}`
	mockNetworkErrorUserReturn = `{"error":"HTTP request failed with error: Unavailable"}`

	mockFoundPassUser       = `{"email":"test@example.com","name":"John Doe","password1":"FoundPassword", "password2":"FoundPassword"}`
	mockFoundPassUserReturn = `{"error":"Password is found in the database."}`

	e    *echo.Echo
	mpwc MockPasswordChecker
	h    Handlers
)

type MockPasswordChecker struct{}

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

func TestMain(m *testing.M) {
	e = echo.New()
	mpwc = MockPasswordChecker{}
	h = NewHandler(mpwc)
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

func TestIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/")

	if assert.NoError(t, h.Index(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestLogin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/login")

	if assert.NoError(t, h.Login(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestRegister(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register")

	if assert.NoError(t, h.Register(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestRegisterPostGood(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(mockGoodUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register")

	if assert.NoError(t, h.RegisterPost(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, mockGoodUserReturn, rec.Body.String())
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
