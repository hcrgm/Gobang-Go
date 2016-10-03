package gobang

import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
	"time"
	"github.com/bitly/go-simplejson"
	"math/rand"
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
		go func() {
			for {
				select {
				case <-ticker.C:
					rooms := simplejson.New()
					roomStatus := simplejson.New()
					for roomId, room := range roomList.rooms {
						roomStatus.Set("playing", room.playing)
						roomStatus.Set("rounds", room.rounds)
						//roomStatus.Set("steps", room.steps)
						roomStatus.Set("steps", rand.Intn(100)) // TODO:test
						roomStatus.Set("watchers", len(room.spectators))
						roomStatus.Set("owner", "Anonymous") //TODO
						rooms.Set(roomId, roomStatus)
					}
					json , err := rooms.Encode()
					if err != nil {
						log.Println(err)
						return
					}
					ws.WriteMessage(websocket.TextMessage, json)
				}
			}
		}()
	}
}

func HandleGameSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: room-joining
		serveWs(NewRoom(), w, r)
	}
}