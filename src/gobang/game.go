package gobang

import (
	"github.com/labstack/echo"
	"net/http"
)
//

func Game(c echo.Context) error {
	return c.Render(http.StatusOK, "game", "")
}