package gobang

import (
	"log"
	"golang.org/x/net/websocket"
)

func Status() websocket.Handler {
	return websocket.Handler(func (ws *websocket.Conn) {
		// test only
		err := websocket.Message.Send(ws, "{\"739617\":{\"owner\":\"test\",\"playing\":true,\"steps\":139,\"rounds\":4,\"watchers\":2}}")
		if err != nil {
			log.Panic(err)
			return
		}
	})
}