package gobang

import (
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var (
	upgrader = websocket.Upgrader{}
)

func HandleStatusSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
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
				roomStatus.Set("owner", "Anonymous") //TODO
				rooms.Set(roomId, roomStatus)
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
					return
				}
			case <-ticker.C:
				if err := update(); err != nil {
					return
				}
			}
		}
	}
}

func HandleGameSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: room-joining
		serveWs(w, r)
	}
}
