package main

import (
	"crypto/sha512"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	rs "github.com/javorszky/go-comments/randomstring"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/masonj88/pwchecker"
)

// Site model definition
type Site struct {
	gorm.Model
	UserID      uint
	Designation string `form:"designation" gorm:"type:varchar(191);not null;unique"`
	Domains     string `form:"domains" gorm:"type:varchar(191)"`
}

// User model definition.
type User struct {
	gorm.Model
	Email          string `json:"email" form:"email" gorm:"type:varchar(191);unique_index:email"`
	PasswordOne    string `form:"password1" gorm:"-" json:"-"`
	PasswordTwo    string `form:"password2" gorm:"-" json:"-"`
	HashedPassword string `json:"passwordHash" gorm:"type:varchar(255)"`
	Sessions       []Session
	Sites          []Site
}

// ResponseError is a generic struct to be turned into JSON in responses.
type ResponseError struct {
	Error string `json:"error"`
}

// PasswordChecker interface to check pw against haveIbeenpwnd API.
type PasswordChecker interface {
	IsPasswordPwnd(string) (bool, error)
}

// PasswordHasher interface to hash and check passwords.
type PasswordHasher interface {
	GenerateFromPassword(string) (string, error)
	ComparePasswordAndHash(string, string) (bool, error)
}

// PwChecker struct implements haveIbeenpwnd API checker.
type PwChecker struct{}

// IsPasswordPwnd is a utility function that checks pw against haveIbeenpwnd API.
func (pw PwChecker) IsPasswordPwnd(password string) (bool, error) {
	pwd, err := pwchecker.CheckForPwnage(password)
	if err != nil {
		return false, err
	}

	return pwd.Pwnd, nil
}

// Handlers struct holds db, passwordhasher, and passwordchecker implementations.
type Handlers struct {
	pwc PasswordChecker
	pwh PasswordHasher
	db  *gorm.DB
}

// BadRegister is a helper struct to return an error and CSRF token.
type BadRegister struct {
	Csrf   interface{}
	Errors []error
}

// NewHandler returns a struct with given implementations.
func NewHandler(pwc PasswordChecker, pwh PasswordHasher, db *gorm.DB) Handlers {
	return Handlers{pwc, pwh, db}
}

// Index handles GET request to /.
func (h *Handlers) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index", "")
}

// Login handles GET request to /login.
func (h *Handlers) Login(c echo.Context) error {
	return c.Render(http.StatusOK, "login", c.Get("csrf"))
}

// LoginPost handles POST request to /login.
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

	sessionID, err := h.setSession(user, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ResponseError{"Something went wrong with setting the session."})
	}

	cookieError := h.setSessionCookie(sessionID, c)
	if cookieError != nil {
		return c.JSON(http.StatusBadRequest, ResponseError{"Something went wrong with setting the session cookie."})
	}

	return cookieError
}

// Logout serves GET to /logout. Destroys cookie
func (h *Handlers) Logout(c echo.Context) error {
	err := h.destroySessionCookie(c)
	if err != nil {
		return c.JSON(http.StatusBadGateway, ResponseError{"Destroying the cookie failed"})
	}

	return err
}

// Register handles GET requests to /register.
func (h *Handlers) Register(c echo.Context) error {
	data := BadRegister{
		c.Get("csrf"),
		nil,
	}
	return c.Render(http.StatusOK, "register", data)
}

// RegisterPost handles POST requests to /register.
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
	}

	return c.JSON(http.StatusOK, u)
}

// Admin serves GET request to /admin
func (h *Handlers) Admin(c echo.Context) error {
	return c.Render(http.StatusOK, "admin", nil)
}

func (h *Handlers) AdminSites(c echo.Context) error {
	return c.Render(http.StatusOK, "adminsites", nil)
}

func (h *Handlers) AdminSitesNew(c echo.Context) error {
	return c.Render(http.StatusOK, "adminnewsite", c.Get("csrf"))
}

// AdminSitesNewPost handles POST /admin/sites/new to add a new entry to the sites
func (h *Handlers) AdminSitesNewPost(c echo.Context) error {
	user, ok := c.Get("model.user").(User)

	if !ok {
		panic("not okay")
	}

	domains, err := json.Marshal(strings.Split(c.FormValue("domains"), "\r\n"))

	if err != nil {
		panic("Can't split thingies")
	}

	domainsString := string(domains)

	site := Site{
		Domains:     domainsString,
		Designation: c.FormValue("designation"),
		UserID:      user.ID,
	}

	if result := h.db.Create(&site); result.Error != nil {
		return c.String(http.StatusBadRequest, "Something failed while saving")
	}

	return c.String(http.StatusCreated, "lel")
}

// ServeJS is handling requests to /:id/js.
func (h *Handlers) ServeJS(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJavaScript)
	return c.Render(http.StatusOK, "client.js", c.Param("id"))
}

// Request is a utility function that helps debug connection details.
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

/*
Internal function to set the session for a user for a given context.

It gets the ID of the user, and IP and User Agent from the context.
Session also has a BeforeCreate hook (see sessions.go) that will
create a uuidv4 as an ID.
*/
func (h *Handlers) setSession(u *User, c echo.Context) (string, error) {

	salt := rs.Generate(16)
	secret := rs.Generate(32)
	source := fmt.Sprintf("%s%s", salt, secret)
	hString := h.hashString(source)

	session := Session{
		UserID:    u.ID,
		IP:        c.Request().RemoteAddr,
		UserAgent: c.Request().UserAgent(),
		Hash:      hString,
	}

	if result := h.db.Create(&session); result.Error != nil {
		return "", result.Error
	}

	return fmt.Sprintf("%s|%s", session.ID, source), nil
}

/*
hashString is a utility function. Calculates the SHA512_256 hash
of a given string, and returns the base64 URL encoded representation
of the source.
*/
func (h *Handlers) hashString(source string) string {
	hasher := sha512.New512_256()
	hasher.Write([]byte(source))
	return b64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

/*
setSessionCookie sets a session cookie with the given value.

Session cookie is valid for 24 hours.
*/
func (h *Handlers) setSessionCookie(value string, c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "gocomments_session"
	cookie.Value = value
	cookie.Expires = time.Now().Add(24 * time.Hour)
	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/admin")
}

// destroySessionCookie sets the expiration of the cookie to 1 year before.
func (h *Handlers) destroySessionCookie(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "gocomments_session"
	cookie.Value = ""
	cookie.Expires = time.Now().AddDate(-1, 0, 0)
	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/login")
}

/*
SessionCheck is a middleware. It takes the context, extracts the cookie,
and then looks up whether there is a session in the database with the
details in the cookie.

If there isn't, it redirects to login page with 302.

If there is, it calls the next middleware.
*/
func (h *Handlers) SessionCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("gocomments_session")
		if err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		session := &Session{}

		splits := strings.Split(cookie.Value, "|")

		hString := h.hashString(splits[1])

		if h.db.Where("id = ?", splits[0]).Where("hash = ?", hString).First(session).RecordNotFound() {
			return c.Redirect(http.StatusFound, "/login")
		}

		user := User{}

		h.db.Where("id = ?", session.UserID).First(&user)

		c.Set("model.user", user)

		return next(c)
	}
}
