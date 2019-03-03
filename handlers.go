package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/masonj88/pwchecker"
	"net/http"
)

type User struct {
	gorm.Model
	Email       string `json:"email" form:"email"`
	Name        string `json:"name" form:"name"`
	PasswordOne string `json:"password1" form:"password1"`
	PasswordTwo string `json:"password2" form:"password2"`
}

type ResponseError struct {
	Error string `json:"error"`
}

type PasswordChecker interface {
	IsPasswordPwnd(string) (bool, error)
}

type PwChecker struct{}

func (pw PwChecker) IsPasswordPwnd(password string) (bool, error) {
	pwd, err := pwchecker.CheckForPwnage(password)
	if err != nil {
		return false, err
	}

	return pwd.Pwnd, nil
}

type Handlers struct {
	pwc PasswordChecker
}

func NewHandler(pwc PasswordChecker) Handlers {
	return Handlers{pwc}
}

func (h *Handlers) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index", "")
}

func (h *Handlers) Login(c echo.Context) error {
	return c.Render(http.StatusOK, "login", "")
}

func (h *Handlers) Register(c echo.Context) error {
	return c.Render(http.StatusOK, "register", c.Get("csrf"))
}

func (h *Handlers) RegisterPost(c echo.Context) (err error) {
	u := new(User)
	if err = c.Bind(u); err != nil {
		return fmt.Errorf("binding user failed")
	}

	if u.PasswordOne != u.PasswordTwo {
		e := ResponseError{Error: "Passwords do not match."}
		return c.JSON(http.StatusUnprocessableEntity, e)
	}

	pwdCheck, err := h.pwc.IsPasswordPwnd(u.PasswordOne)

	// Something went wrong while checking the API
	if err != nil {
		e := ResponseError{Error: err.Error()}
		return c.JSON(http.StatusBadGateway, e)
	}

	if pwdCheck {
		e := ResponseError{Error: "Password is found in the database."}
		return c.JSON(http.StatusUnprocessableEntity, e)
	}

	return c.JSON(http.StatusOK, u)
}

func (h *Handlers) ServeJS(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJavaScript)
	return c.Render(http.StatusOK, "client.js", c.Param("id"))
}

func (h *Handlers) Request(c echo.Context) error {
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
}
