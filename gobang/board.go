package gobang

const (
	EMPTY = 0
	BLACK = 1
	WHITE = 2
	SIZE  = 15
)

type Board struct {
	lastStepX int
	lastStepY int
	cells     [SIZE][SIZE]int
}

func NewBoard() *Board {
	var cell [SIZE][SIZE]int
	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			cell[i][j] = EMPTY
		}
	}
	return &Board{
		lastStepX: -1,
		lastStepY: -1,
		cells:     cell,
	}
}

func (board *Board) getTimes(cx, cy, dx, dy, c int) int {
	if c == EMPTY {
		return 0
	}
	if dx == 0 && dy == 0 {
		return 0
	}
	times := 0
	for i := 1; i <= 5; i++ {
		nx := cx + (dx * i)
		ny := cy + (dy * i)
		if nx < 0 || ny < 0 || nx >= len(board.cells) || ny >= len(board.cells[0]) {
			continue
		}
		nc := board.cells[nx][ny]
		if nc == EMPTY || c != nc {
			break
		}
		times++
	}
	return times
}

func (board *Board) checkWin(x, y, d int, color int) bool {
	if d == EMPTY {
		return false
	}
	if (board.getTimes(x, y, 0, 1, d)+board.getTimes(x, y, 0, -1, d)) >= 4 || (board.getTimes(x, y, 1, 0, d)+board.getTimes(x, y, -1, 0, d)) >= 4 || (board.getTimes(x, y, 1, 1, d)+board.getTimes(x, y, -1, -1, d)) >= 4 || (board.getTimes(x, y, 1, -1, d)+board.getTimes(x, y, -1, 1, d)) >= 4 {
		if color == BLACK {
			return d == BLACK
		} else if color == WHITE {
			return d == WHITE
		}
	}
	return false
}

func CheckData(data int) bool {
	if data == BLACK || data == WHITE {
		return true
	}
	return false
}

func GetColor(data int) string {
	if data == BLACK {
		return "BLACK"
	} else if data == WHITE {
		return "WHITE"
	} else {
		return "EMPTY"
	}
}
