package handlers

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"net/http"
)

type User struct {
	gorm.Model
	Email        string `json:"email" form:"email"`
	Name         string `json:"name" form:"name"`
	PasswordHash string `json:"password" form:"password"`
}

func Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index", "")
}

func Login(c echo.Context) error {
	return c.Render(http.StatusOK, "login", "")
}

func Register(c echo.Context) error {
	return c.Render(http.StatusOK, "register", "")
}

func RegisterPost(c echo.Context) (err error) {
	fmt.Println("this is registerpost")
	u := new(User)
	if err = c.Bind(u); err != nil {
		return fmt.Errorf("binding user failed")
	}
	return c.JSON(http.StatusOK, u)
}

func ServeJS(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJavaScript)
	return c.Render(http.StatusOK, "client.js", c.Param("id"))
}

func Request(c echo.Context) error {
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
