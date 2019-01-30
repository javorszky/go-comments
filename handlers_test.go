package main

import (
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	serveJS = `console.log("The id of the thing requesting this was 44");`
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
