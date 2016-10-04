package gobang

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/random"
	"log"
	"math/rand"
	"net/http"
	"time"
	"golang.org/x/net/html"
	"github.com/kataras/go-sessions"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	name string
	room *Room
	ws   *websocket.Conn
	send chan []byte
}

func (c *Client) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *Client) readPump() {
	// TODO: chat
	defer func() {
		c.room.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		fmt.Println("mes:" + string(message))
		c.room.broadcastAll <- []byte(html.EscapeString(string(message))) // block javascript, etc
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

type RoomList struct {
	rooms map[string]*Room
}

var roomList = &RoomList{
	rooms: make(map[string]*Room),
}

type Room struct {
	roomId       string
	playerBlack  *Client
	playerWhite  *Client
	spectators   map[*Client]bool
	playing      bool
	holding      bool // true for black, false for white
	steps        int
	rounds       int
	broadcastAll chan []byte
	register     chan *Client
	unregister   chan *Client
}

func NewRoom() *Room {
	room := &Room{
		roomId:       random.String(8),
		playerBlack:  nil,
		playerWhite:  nil,
		spectators:   make(map[*Client]bool),
		playing:      false,
		holding:      false,
		steps:        0,
		rounds:       0,
		broadcastAll: make(chan []byte, maxMessageSize),
		register:     make(chan *Client, maxMessageSize),
		unregister:   make(chan *Client, maxMessageSize),
	}
	roomList.rooms[room.roomId] = room
	return room
}

func (room *Room) startGame(restart bool) {
	room.steps = 0
	if !restart {
		room.sendToBlack([]byte("start:black"))
		room.sendToWhite([]byte("start:white"))
		if room.holding {
			room.sendToAll([]byte("status:black:Holding..."))
			room.sendToAll([]byte("status:white:Waiting..."))
		} else {
			room.sendToAll([]byte("status:black:Waiting..."))
			room.sendToAll([]byte("status:white:Holding..."))
		}
	}
	// TODO: restart
	// TODO: clear cells
	//room.broadcastAll<-[]byte("clear")
	//room.playing = true
	//room.rounds++
}

func (room *Room) sendToBlack(message []byte) {
	if room.playerBlack != nil {
		room.playerBlack.write(websocket.TextMessage, message)
	}
}

func (room *Room) sendToWhite(message []byte) {
	if room.playerWhite != nil {
		room.playerWhite.write(websocket.TextMessage, message)
	}
}

func (room *Room) sendToSpectators(message []byte) {
	for spectator := range room.spectators {
		spectator.send <- message
	}
}

func (room *Room) sendToAll(message []byte) {
	room.sendToBlack(message)
	room.sendToWhite(message)
	room.sendToSpectators(message)
}

func (room *Room) onQuit(client *Client) (deleteRoom bool) {
	if client == room.playerBlack {
		// Black left the game
		fmt.Println("Black left")
		room.playerBlack = nil
		room.broadcastAll <- []byte("chat:System:Black left")
	} else if client == room.playerWhite {
		// White left the game
		fmt.Println("White left")
		room.playerWhite = nil
		room.broadcastAll <- []byte("chat:System:White left")
	} else if _, ok := room.spectators[client]; ok {
		delete(room.spectators, client)
	}
	close(client.send)
	if room.playerBlack == nil && room.playerWhite == nil {
		deleteRoom = true
	}
	return
}

func (room *Room) onJoin(client *Client) {
	if room.playing {
		room.spectators[client] = true // Spectator joined
		// Handle spectator
		client.write(websocket.TextMessage, []byte("start:spectator"))
		blackStatus := ""
		whiteStatus := ""
		if room.holding {
			blackStatus = "Holding..."
			whiteStatus = "Waiting..."
		} else {
			blackStatus = "Waiting..."
			whiteStatus = "Holding..."
		}
		client.write(websocket.TextMessage, []byte("status:black:" + blackStatus))
		client.write(websocket.TextMessage, []byte("status:white:" + whiteStatus))
		client.write(websocket.TextMessage, []byte("join:black:" + room.playerBlack.name))
		client.write(websocket.TextMessage, []byte("join:white:" + room.playerWhite.name))
		client.write(websocket.TextMessage, []byte("join:spectator:" + client.name))
	} else if room.playerBlack == nil && room.playerWhite == nil {
		if isBlack := rand.Int31n(2); isBlack == 1 {
			room.playerBlack = client
			room.holding = true
		} else {
			room.playerWhite = client
		}
		client.send <- []byte("room:" + room.roomId)
	} else {
		if room.playerBlack == nil {
			room.playerBlack = client
			room.broadcastAll <- []byte("join:black:" + client.name)
			client.send <- []byte("join:white:" + room.playerWhite.name)
		} else if room.playerWhite == nil {
			room.playerWhite = client
			room.broadcastAll <- []byte("join:white:" + client.name)
			client.send <- []byte("join:black:" + room.playerBlack.name)
		}
		room.startGame(false)
	}
}

func (room *Room) run() {
	for {
		select {
		case client := <-room.register:
			fmt.Println("Register a user")
			room.onJoin(client)
		case client := <-room.unregister:
			// Check if we can delete the room
			if room.onQuit(client) {
				fmt.Println("deleting room")
				delete(roomList.rooms, room.roomId)
			}
		case message := <-room.broadcastAll:
			room.sendToAll(message)
		}
	}
}
func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	sess := sessions.Start(w, r)
	var room *Room = nil
	if create := sess.GetString("create"); create == "true" {
		room = NewRoom()
		sess.Set("create", false)
	} else if roomId := sess.GetString("roomId"); roomId != "" {
		if getroom, ok := roomList.rooms[roomId]; ok {
			room = getroom
			sess.Set("roomId", "")
		} else {
			ws.WriteMessage(websocket.TextMessage, []byte("err:Can't join the room:" + roomId))
			ws.Close()
			return
		}
	} else {
		ws.WriteMessage(websocket.TextMessage, []byte("err:Internal Server Error"))
		ws.Close()
		return
	}
	client := &Client{name: "Anonymous", room: room, ws: ws, send: make(chan []byte, maxMessageSize)}
	client.room.register <- client
	go room.run()
	go client.writePump()
	client.readPump()
}
