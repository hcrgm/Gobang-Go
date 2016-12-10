package gobang

import (
	"fmt"
	"github.com/kataras/go-sessions"
	"github.com/labstack/echo"
	"net/http"
	"strings"
)

func Game(c echo.Context) error {
	create := false
	roomId := ""
	queryString := c.Request().URL.RawQuery
	w := c.Response().Writer()
	r := c.Request()
	sess := sessions.Start(w, r)
	if strings.EqualFold(queryString, "closed") {
		return c.Redirect(http.StatusMovedPermanently, "index.html")
	}
	if strings.EqualFold(queryString, "create") {
		create = true
	} else if len(queryString) != 0 {
		if room, ok := roomList.rooms[queryString]; ok {
			roomId = room.roomId
			sess.Set("roomId", roomId)
		} else {
			return c.HTML(http.StatusNotFound, fmt.Sprintf(`<script>alert("Can not found room id %s!");location.href="index.html";</script>`, queryString))
		}
	} else {
		create = true
	}
	if create {
		sess.Set("create", "true")
	}
	name := "Anonymous"
	if len(sess.GetString("name")) != 0 {
		name = sess.GetString("name")
	}
	return c.Render(http.StatusOK, "game", struct {
		Create   bool
		RoomId   string
		Username string
	}{Create: create, RoomId: roomId, Username: name})
}
