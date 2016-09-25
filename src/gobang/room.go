package gobang

// Not implemented yet

const EMPTY = 0
const BLACK = 1
const WHITE  = 2

const NONE = 0
const PLAYER1 = 1
const PLAYER2 = 2

type Room struct {
	RoomID int
	Player1 User
	Player2 User
	Spectators map[string]*User
	Playing bool
	UndoRequest int
	TurnToBlack bool
	Stpes int
	Rounds int
	Cells [15][15]int
	LastStepX int
	LastStepY int
}

