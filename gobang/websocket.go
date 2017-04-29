package gobang

import (
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"log"
	"time"
)

var (
	upgrader = websocket.Upgrader{}
)

func HandleStatusSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(3 * time.Second)
	timer := time.NewTimer(time.Second)
	update := func() error {
		rooms := simplejson.New()
		roomStatus := simplejson.New()
		for roomId, room := range roomList.rooms {
			roomStatus.Set("playing", room.playing)
			roomStatus.Set("rounds", room.rounds)
			roomStatus.Set("steps", room.steps)
			roomStatus.Set("watchers", len(room.spectators))
			roomStatus.Set("owner", room.owner.name)
			rooms.Set(roomId, roomStatus)
			roomStatus = simplejson.New()
		}
		json, err := rooms.Encode()
		if err != nil {
			log.Println(err)
			json = []byte("{}")
		}
		return ws.WriteMessage(websocket.TextMessage, json)
	}
	defer func() {
		ticker.Stop()
		timer.Stop()
		ws.Close()
	}()
	for {
		select {
		case <-timer.C:
			if err := update(); err != nil {
				return err
			}
		case <-ticker.C:
			if err := update(); err != nil {
				return err
			}
		}
	}
}

func HandleGameSocket(c echo.Context) error {
	return serveWs(c.Response(), c.Request())
}
