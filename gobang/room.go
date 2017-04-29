package gobang

import (
	"github.com/gorilla/websocket"
	"github.com/kataras/go-sessions"
	"github.com/labstack/gommon/random"
	"golang.org/x/net/html"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	name  string
	room  *Room
	ws    *websocket.Conn
	send  chan []byte
	mutex sync.Mutex
}

func NewClient(name string, room *Room, ws *websocket.Conn) *Client {
	return &Client{name: name, room: room, ws: ws, send: make(chan []byte)}
}

func (c *Client) write(mt int, payload []byte) error {
	c.mutex.Lock()
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	defer c.mutex.Unlock()
	return c.ws.WriteMessage(mt, payload)
}

func (c *Client) writeTextMessage(message []byte) error {
	return c.write(websocket.TextMessage, message)
}

func (c *Client) readPump() {
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
		log.Println("mes:" + string(message))
		c.room.onMessage(message, c)
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
			err := c.writeTextMessage(message)
			if err != nil {
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
	owner        *Client
	playerBlack  *Client
	playerWhite  *Client
	spectators   map[*Client]bool
	playing      bool
	holding      bool // true for black, false for white
	board        *Board
	steps        int
	rounds       int
	undoRequest  int // zero for no any undo request
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
		board:        NewBoard(),
		steps:        0,
		rounds:       0,
		broadcastAll: make(chan []byte),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
	}
	roomList.rooms[room.roomId] = room
	return room
}

func (room *Room) startGame(restart bool) {
	if !room.canStart() {
		return
	}
	room.steps = 0
	if !restart {
		if room.holding {
			room.sendToBlack([]byte("start:black"))
			room.sendToWhite([]byte("start:white"))
			room.sendToAll([]byte("status:black:Holding..."))
			room.sendToAll([]byte("status:white:Waiting..."))
		} else {
			room.sendToBlack([]byte("start:white"))
			room.sendToWhite([]byte("start:black"))
			room.sendToAll([]byte("status:black:Waiting..."))
			room.sendToAll([]byte("status:white:Holding..."))
		}
	}
	room.undoRequest = 0
	if room.holding {
		room.sendToAll([]byte("turn:BLACK"))
	} else {
		room.sendToAll([]byte("turn:WHITE"))
	}
	room.sendToAll([]byte("clear"))
	room.updateAllMap(EMPTY, room.playerBlack, room.playerWhite)
	room.sendToAll([]byte("clear"))
	room.playing = true
	room.rounds++
}

func (room *Room) sendToBlack(message []byte) {
	if room.playerBlack != nil {
		room.playerBlack.writeTextMessage(message)
	}
}

func (room *Room) sendToWhite(message []byte) {
	if room.playerWhite != nil {
		room.playerWhite.writeTextMessage(message)
	}
}

func (room *Room) sendToSpectators(message []byte) {
	for spectator := range room.spectators {
		spectator.writeTextMessage(message)
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
		room.playerBlack = nil
		room.gameOver("Black left the game", false)
		if room.playerWhite != nil {
			room.playerWhite.write(websocket.CloseMessage, []byte("closesocket"))
		}
		room.playerWhite = nil
		room.closeSpectators()
	} else if client == room.playerWhite {
		// White left the game
		room.playerWhite = nil
		room.gameOver("White left the game", false)
		if room.playerBlack != nil {
			room.playerBlack.write(websocket.CloseMessage, []byte("closesocket"))
		}
		room.playerBlack = nil
		room.closeSpectators()
	} else if _, ok := room.spectators[client]; ok {
		delete(room.spectators, client)
	}
	close(client.send)
	if room.playerBlack == nil && room.playerWhite == nil {
		deleteRoom = true
	}
	return
}

func (room *Room) closeSpectators() {
	for spectator := range room.spectators {
		spectator.write(websocket.CloseMessage, []byte("closesocket"))
		delete(room.spectators, spectator)
	}
	room.spectators = nil
}

func (room *Room) update(x, y int) {
	room.updateToAll(x, y, room.board.cells[x][y])
}

func (room *Room) updateToAll(x, y, data int) {
	room.sendToAll([]byte("update:" + strconv.Itoa(x) + ":" + strconv.Itoa(y) + ":" + strconv.Itoa(data)))
}

func (room *Room) updateToBlack(x, y int) {
	room.sendToBlack([]byte("update:" + strconv.Itoa(x) + ":" + strconv.Itoa(y) + ":" + strconv.Itoa(room.board.cells[x][y])))
}

func (room *Room) updateToWhite(x, y int) {
	room.sendToBlack([]byte("update:" + strconv.Itoa(x) + ":" + strconv.Itoa(y) + ":" + strconv.Itoa(room.board.cells[x][y])))
}

func (room *Room) updateAllMap(nc int, clients ...*Client) {
	for x := 0; x < len(room.board.cells); x++ {
		for y := 0; y < len(room.board.cells[0]); y++ {
			if nc != -1 {
				room.board.cells[x][y] = nc
			}
			for _, client := range clients {
				if err := client.writeTextMessage([]byte("update:" + strconv.Itoa(x) + ":" + strconv.Itoa(y) + ":" + strconv.Itoa(room.board.cells[x][y]))); err != nil {
					return
				}
			}
		}
	}
}

func (room *Room) isSpectator(client *Client) bool {
	return client == room.playerBlack && client == room.playerWhite
}

func (room *Room) canStart() bool {
	return !room.playing && room.playerBlack != nil && room.playerWhite != nil
}

func (room *Room) canUndo(holder *Client) bool {
	if room.isSpectator(holder) {
		return false
	}
	if room.board.lastStepX == -1 || room.board.lastStepY == -1 {
		return false
	}
	if (holder == room.playerBlack && !room.holding) || (holder == room.playerWhite && room.holding) {
		return true
	}
	return false
}

func (room *Room) gameOver(message string, canRestart bool) {
	if !room.playing {
		return
	}
	room.playing = false
	room.sendToAll([]byte("gameover:" + message))
	if canRestart && room.canStart() {
		room.startGame(true)
	}
}

func (room *Room) onJoin(client *Client) {
	if room.playing {
		room.spectators[client] = true // Spectator joined
		client.writeTextMessage([]byte("start:spectator"))
		room.updateAllMap(-1, client)
		blackStatus := ""
		whiteStatus := ""
		if room.holding {
			blackStatus = "Holding..."
			whiteStatus = "Waiting..."
		} else {
			blackStatus = "Waiting..."
			whiteStatus = "Holding..."
		}
		client.writeTextMessage([]byte("status:black:" + blackStatus))
		client.writeTextMessage([]byte("status:white:" + whiteStatus))
		client.writeTextMessage([]byte("join:black:" + room.playerBlack.name))
		client.writeTextMessage([]byte("join:white:" + room.playerWhite.name))
		room.sendToAll([]byte("join:spectator:" + client.name))
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
			room.holding = true
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

func (room *Room) onMessage(message []byte, client *Client) {
	message = []byte(html.EscapeString(string(message))) // block javascript, etc
	slices := strings.Split(string(message), ":")
	if len(slices) == 0 {
		return
	}
	if slices[0] == "chat" {
		if len(slices) < 3 {
			return
		}
		if slices[1] != client.name {
			return
		}
		chatMessage := ""
		for i := 2; i < len(slices); i++ {
			chatMessage += slices[i]
		}
		if len(chatMessage) > 50 {
			client.writeTextMessage([]byte("chat:System:Too long!"))
			return
		}
		prefix := "[S]"
		if client == room.playerBlack {
			prefix = "[B]"
		} else if client == room.playerWhite {
			prefix = "[W]"
		}
		room.broadcastAll <- []byte("chat:" + prefix + client.name + ":" + chatMessage)
	}
	if client == room.playerBlack || client == room.playerWhite {
		switch slices[0] {
		case "update":
			if len(slices) < 4 {
				log.Println("Bad update message")
				return
			}
			x, err1 := strconv.Atoi(slices[1])
			y, err2 := strconv.Atoi(slices[2])
			data, err3 := strconv.Atoi(slices[3])
			if err1 != nil || err2 != nil || err3 != nil {
				log.Println("Parse error")
				return
			}
			if x < 0 || x > 14 || y < 0 || y > 14 || !CheckData(data) {
				log.Println("Bad coordinate")
				return
			}
			if room.board.cells[x][y] != EMPTY {
				log.Println("Cannot refill the cell")
				client.writeTextMessage([]byte("update:" + strconv.Itoa(x) + ":" + strconv.Itoa(y) + ":" + strconv.Itoa(room.board.cells[x][y])))
				return
			}
			color := room.board.cells[x][y]
			room.board.lastStepX = x
			room.board.lastStepY = y
			if room.holding {
				if client == room.playerBlack {
					color = BLACK
				} else {
					room.updateToWhite(x, y)
					return
				}
			} else {
				if client == room.playerWhite {
					color = WHITE
				} else {
					room.updateToBlack(x, y)
					return
				}
			}
			room.steps++
			room.board.cells[x][y] = color
			room.board.lastStepX = x
			room.board.lastStepY = y
			room.update(x, y)
			room.holding = !room.holding
			turnTo := ""
			if room.holding {
				turnTo = "BLACK"
			} else {
				turnTo = "WHITE"
			}
			room.sendToAll([]byte("turn:" + turnTo))
			if room.board.checkWin(x, y, room.board.cells[x][y], color) {
				room.gameOver(GetColor(color)+" win!", true)
			}
		case "status":
			room.sendToAll(message)
		case "undo":
			if len(slices) < 2 {
				return
			}
			switch slices[1] {
			case "request":
				if room.canUndo(client) {
					if room.undoRequest != 0 {
						return
					}
					if client == room.playerBlack {
						room.undoRequest = BLACK
					} else {
						room.undoRequest = WHITE
					}
					room.sendToAll([]byte("chat:System:" + strings.Title(strings.ToLower(GetColor(room.undoRequest))) + " wants to undo one step"))
					room.sendToAll([]byte("undo:request:" + strings.ToLower(GetColor(room.undoRequest))))
				}
			case "accept":
				if len(slices) < 2 {
					return
				}
				if room.isSpectator(client) {
					return
				}
				if (client == room.playerBlack && room.undoRequest != WHITE) || (client == room.playerWhite && room.undoRequest != BLACK) {
					return
				}
				if room.board.lastStepX == -1 || room.board.lastStepY == -1 {
					return
				}
				room.board.cells[room.board.lastStepX][room.board.lastStepY] = EMPTY
				room.update(room.board.lastStepX, room.board.lastStepY)
				color := ""
				if client == room.playerBlack {
					room.sendToWhite([]byte("undo:accept"))
					color = GetColor(WHITE)
				} else {
					room.sendToBlack([]byte("undo:accept"))
					color = GetColor(BLACK)
				}
				color = strings.Title(strings.ToLower(color))
				room.sendToAll([]byte("chat:System:" + color + " undid one step..."))
				room.sendToSpectators([]byte("undo:accept:" + color))
				room.holding = !room.holding
				room.sendToAll([]byte("turn:" + strings.ToUpper(color) + ":n"))
				room.undoRequest = 0
			case "deny":
				if len(slices) < 2 {
					return
				}
				if room.isSpectator(client) {
					return
				}
				if (client == room.playerBlack && room.undoRequest != WHITE) || (client == room.playerWhite && room.undoRequest != BLACK) {
					return
				}
				if room.board.lastStepX == -1 || room.board.lastStepY == -1 {
					return
				}
				color := ""
				if client == room.playerBlack {
					room.sendToWhite([]byte("undo:deny"))
					color = "White"
				} else {
					room.sendToBlack([]byte("undo:deny"))
					color = "Black"
				}
				room.sendToAll([]byte("chat:System:Undo denied..."))
				room.sendToSpectators([]byte("undo:deny:" + color))
				room.undoRequest = 0
			}
		}
	}
}

func (room *Room) run() {
	for {
		select {
		case client := <-room.register:
			room.onJoin(client)
		case client := <-room.unregister:
			// Check if we can delete the room
			if room.onQuit(client) {
				delete(roomList.rooms, room.roomId)
			}
		case message := <-room.broadcastAll:
			room.sendToAll(message)
		}
	}
}
func serveWs(w http.ResponseWriter, r *http.Request) error {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	sess := sessions.Start(w, r)
	var room *Room = nil
	name := "Anonymous"
	if len(sess.GetString("name")) != 0 {
		name = sess.GetString("name")
	}
	if create := sess.GetString("create"); create == "true" {
		room = NewRoom()
		room.owner = NewClient(name, room, ws)
		sess.Delete("create")
	} else if roomId := sess.GetString("roomId"); roomId != "" {
		if getroom, ok := roomList.rooms[roomId]; ok {
			room = getroom
			sess.Delete("roomId")
		} else {
			ws.WriteMessage(websocket.TextMessage, []byte("err:Can't join the room:"+roomId))
			ws.Close()
			return nil
		}
	} else {
		ws.WriteMessage(websocket.TextMessage, []byte("err:Internal Server Error"))
		ws.Close()
		return nil
	}
	client := NewClient(name, room, ws)
	go room.run()
	client.room.register <- client
	go client.writePump()
	client.readPump()
	return nil
}
