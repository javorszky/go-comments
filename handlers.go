package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/masonj88/pwchecker"
	"net/http"
	"regexp"
)

type (
	User struct {
		gorm.Model
		Email          string `json:"email" form:"email" gorm:"type:varchar(191);unique_index:email" validate:"required,email"`
		PasswordOne    string `form:"password1" gorm:"-" json:"-" validate:"required,min=8"`
		PasswordTwo    string `form:"password2" gorm:"-" json:"-" validate:"omitempty,eqfield=PasswordOne"`
		HashedPassword string `json:"passwordHash" gorm:"type:varchar(255)"`
	}

	ResponseError struct {
		Error string `json:"error"`
	}

	PasswordChecker interface {
		IsPasswordPwnd(string) (bool, error)
	}

	PasswordHasher interface {
		GenerateFromPassword(string) (string, error)
		ComparePasswordAndHash(string, string) (bool, error)
	}

	PwChecker struct{}
)

func (pw PwChecker) IsPasswordPwnd(password string) (bool, error) {
	pwd, err := pwchecker.CheckForPwnage(password)
	if err != nil {
		return false, err
	}

	return pwd.Pwnd, nil
}

type Handlers struct {
	pwc PasswordChecker
	pwh PasswordHasher
	db  *gorm.DB
}

type BadRegister struct {
	Csrf   interface{}
	Errors []error
}

func NewHandler(pwc PasswordChecker, pwh PasswordHasher, db *gorm.DB) Handlers {
	return Handlers{pwc, pwh, db}
}

func (h *Handlers) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index", "")
}

func (h *Handlers) Login(c echo.Context) error {
	return c.Render(http.StatusOK, "login", c.Get("csrf"))
}

func (h *Handlers) LoginPost(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Check that passed email is actually an email. Snippet taken from
	// https://www.alexedwards.net/blog/validation-snippets-for-go#email-validation
	rxEmail := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if len(email) > 254 || !rxEmail.MatchString(email) {
		return c.JSON(http.StatusBadRequest, ResponseError{"Passed email is not an email format."})
	}

	// Check that there's a non-empty password
	if password == "" {
		return c.JSON(http.StatusBadRequest, ResponseError{"Passed password is empty."})
	}

	user := &User{}

	if h.db.Where("email = ?", email).First(user).RecordNotFound() {
		return c.JSON(http.StatusNotFound, ResponseError{"No user by that email address."})
	}

	match, err := h.pwh.ComparePasswordAndHash(password, user.HashedPassword)

	if err != nil {
		return c.JSON(http.StatusBadRequest, ResponseError{"Checking passwords failed."})
	}

	if !match {
		return c.JSON(http.StatusUnauthorized, ResponseError{"Passwords do not match."})
	}

	return c.JSON(http.StatusOK, ResponseError{"Passwords match."})
}

func (h *Handlers) Register(c echo.Context) error {
	data := BadRegister{
		c.Get("csrf"),
		nil,
	}
	return c.Render(http.StatusOK, "register", data)
}

func (h *Handlers) RegisterPost(c echo.Context) (err error) {
	u := new(User)

	if err = c.Bind(u); err != nil {
		return fmt.Errorf("binding user failed")
	}

	if u.PasswordOne == "" || u.PasswordTwo == "" {
		e := ResponseError{Error: "No password was passed."}
		return c.JSON(http.StatusUnprocessableEntity, e)
	}

	if u.PasswordOne != u.PasswordTwo {
		e := ResponseError{Error: "Passwords do not match."}
		return c.JSON(http.StatusUnprocessableEntity, e)
	}

	hashedPassword, err := h.pwh.GenerateFromPassword(u.PasswordOne)

	if err != nil {
		e := ResponseError{Error: err.Error()}

		return c.JSON(http.StatusBadGateway, e)
	}

	u.HashedPassword = hashedPassword

	if result := h.db.Create(&u); result.Error != nil {
		data := BadRegister{
			c.Get("csrf"),
			result.GetErrors(),
		}
		return c.JSON(http.StatusConflict, data)

		//return c.Render(http.StatusBadRequest, "register", data)
	}

	//pwdCheck, err := h.pwc.IsPasswordPwnd(u.PasswordOne)
	//
	//// Something went wrong while checking the API
	//if err != nil {
	//	e := ResponseError{Error: err.Error()}
	//	return c.JSON(http.StatusBadGateway, e)
	//}
	//
	//if pwdCheck {
	//	e := ResponseError{Error: "Password is found in the database."}
	//	return c.JSON(http.StatusUnprocessableEntity, e)
	//}

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
