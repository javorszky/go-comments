package main

import (
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	serveJS        = `console.log("The id of the thing requesting this was 44");`
	mockUser       = `{"email":"test@example.com","name":"John Doe","password1":"somepassword", "password2":"somepassword"}`
	mockUserReturn = `{"ID":0,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","DeletedAt":null,"email":"test@example.com","name":"John Doe","password1":"somepassword","password2":"somepassword"}`

	mockBadUser       = `{"email":"test@example.com","name":"John Doe","password1":"somepassword", "password2":"someotherpass"}`
	mockBadUserReturn = `{"error":"Passwords do not match."}`
)

func TestServeJS(t *testing.T) {
	e := echo.New()
	SetRenderer(e)

	req := httptest.NewRequest(http.MethodGet, "/44/js", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/:id/js")
	c.SetParamNames("id")
	c.SetParamValues("44")

	if assert.NoError(t, ServeJS(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, serveJS, rec.Body.String())
	}
}

func TestIndex(t *testing.T) {
	e := echo.New()
	SetRenderer(e)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/")

	if assert.NoError(t, Index(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestLogin(t *testing.T) {
	e := echo.New()
	SetRenderer(e)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/login")

	if assert.NoError(t, Login(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestRegister(t *testing.T) {
	e := echo.New()
	SetRenderer(e)

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register")

	if assert.NoError(t, Register(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestRegisterPostGood(t *testing.T) {
	e := echo.New()
	SetRenderer(e)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(mockUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register")

	if assert.NoError(t, RegisterPost(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, mockUserReturn, rec.Body.String())
	}
}

func TestRegisterPostBad(t *testing.T) {
	e := echo.New()
	SetRenderer(e)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(mockBadUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/register")

	if assert.NoError(t, RegisterPost(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, mockBadUserReturn, rec.Body.String())
	}
}
