package gobang

import (
	"github.com/labstack/echo"
	"net/http"
	"strings"
	"fmt"
)

func Game(c echo.Context) error {
	create := false
	roomId := ""
	queryString := c.Request().URL().QueryString()
	if strings.EqualFold(queryString, "closed") {
		return c.Redirect(http.StatusMovedPermanently, "index.html")
	}
	if strings.EqualFold(queryString, "create") {
		create = true
	} else if len(queryString) != 0 {
		if room, ok := roomList.rooms[queryString]; ok {
			roomId = room.roomId
		} else {
			return c.HTML(http.StatusNotFound, fmt.Sprintf(`<script>alert("Can not found room id %s!");location.href="index.html";</script>`, queryString))
		}
	} else {
		create = true
	}
	return c.Render(http.StatusOK, "game", struct {
		Create bool
		RoomId string
		Username string
	}{Create: create, RoomId: roomId, Username: "Anonymous"})
}