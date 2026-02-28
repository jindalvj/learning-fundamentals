package chessgame

// Color represents the color of a chess piece or player
type Color int

const (
	White Color = iota
	Black
)

func (c Color) String() string {
	if c == White {
		return "White"
	}
	return "Black"
}

func (c Color) Opponent() Color {
	if c == White {
		return Black
	}
	return White
}

// PieceType represents the type of chess piece
type PieceType int

const (
	King PieceType = iota
	Queen
	Rook
	Bishop
	Knight
	Pawn
)

func (p PieceType) String() string {
	switch p {
	case King:
		return "King"
	case Queen:
		return "Queen"
	case Rook:
		return "Rook"
	case Bishop:
		return "Bishop"
	case Knight:
		return "Knight"
	case Pawn:
		return "Pawn"
	}
	return "Unknown"
}

// Position represents a cell on the board (0-indexed)
type Position struct {
	Row int // 0 = rank 1 (white's back rank)
	Col int // 0 = file a
}

func (p Position) IsValid() bool {
	return p.Row >= 0 && p.Row < 8 && p.Col >= 0 && p.Col < 8
}

func (p Position) String() string {
	return string(rune('a'+p.Col)) + string(rune('1'+p.Row))
}

// GameStatus represents the current status of the game
type GameStatus int

const (
	InProgress GameStatus = iota
	Check
	Checkmate
	Stalemate
)
