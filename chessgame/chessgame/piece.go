package chessgame

// Piece represents a chess piece
type Piece struct {
	Type     PieceType
	Color    Color
	Position Position
	HasMoved bool // used for castling and pawn double-move logic
}

func NewPiece(t PieceType, c Color, pos Position) *Piece {
	return &Piece{Type: t, Color: c, Position: pos}
}

// Symbol returns a Unicode chess symbol for display
func (p *Piece) Symbol() string {
	symbols := map[Color]map[PieceType]string{
		White: {King: "♔", Queen: "♕", Rook: "♖", Bishop: "♗", Knight: "♘", Pawn: "♙"},
		Black: {King: "♚", Queen: "♛", Rook: "♜", Bishop: "♝", Knight: "♞", Pawn: "♟"},
	}
	return symbols[p.Color][p.Type]
}

// PotentialMoves returns all squares this piece could move to, ignoring check constraints.
// The board is passed for context (captures, en passant, castling).
func (p *Piece) PotentialMoves(b *Board) []Position {
	switch p.Type {
	case King:
		return p.kingMoves(b)
	case Queen:
		return p.slidingMoves(b, [][]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}})
	case Rook:
		return p.slidingMoves(b, [][]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}})
	case Bishop:
		return p.slidingMoves(b, [][]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}})
	case Knight:
		return p.knightMoves(b)
	case Pawn:
		return p.pawnMoves(b)
	}
	return nil
}

func (p *Piece) slidingMoves(b *Board, directions [][]int) []Position {
	var moves []Position
	for _, d := range directions {
		for i := 1; i < 8; i++ {
			pos := Position{Row: p.Position.Row + d[0]*i, Col: p.Position.Col + d[1]*i}
			if !pos.IsValid() {
				break
			}
			target := b.PieceAt(pos)
			if target == nil {
				moves = append(moves, pos)
			} else {
				if target.Color != p.Color {
					moves = append(moves, pos) // capture
				}
				break // blocked
			}
		}
	}
	return moves
}

func (p *Piece) knightMoves(b *Board) []Position {
	offsets := [][]int{{-2, -1}, {-2, 1}, {2, -1}, {2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}}
	var moves []Position
	for _, o := range offsets {
		pos := Position{Row: p.Position.Row + o[0], Col: p.Position.Col + o[1]}
		if pos.IsValid() {
			target := b.PieceAt(pos)
			if target == nil || target.Color != p.Color {
				moves = append(moves, pos)
			}
		}
	}
	return moves
}

func (p *Piece) kingMoves(b *Board) []Position {
	offsets := [][]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}
	var moves []Position
	for _, o := range offsets {
		pos := Position{Row: p.Position.Row + o[0], Col: p.Position.Col + o[1]}
		if pos.IsValid() {
			target := b.PieceAt(pos)
			if target == nil || target.Color != p.Color {
				moves = append(moves, pos)
			}
		}
	}

	// Castling
	if !p.HasMoved && !b.IsInCheck(p.Color) {
		row := p.Position.Row
		// King-side castling
		kRook := b.PieceAt(Position{Row: row, Col: 7})
		if kRook != nil && kRook.Type == Rook && kRook.Color == p.Color && !kRook.HasMoved {
			if b.PieceAt(Position{Row: row, Col: 5}) == nil &&
				b.PieceAt(Position{Row: row, Col: 6}) == nil &&
				!b.IsSquareAttacked(Position{Row: row, Col: 5}, p.Color.Opponent()) &&
				!b.IsSquareAttacked(Position{Row: row, Col: 6}, p.Color.Opponent()) {
				moves = append(moves, Position{Row: row, Col: 6})
			}
		}
		// Queen-side castling
		qRook := b.PieceAt(Position{Row: row, Col: 0})
		if qRook != nil && qRook.Type == Rook && qRook.Color == p.Color && !qRook.HasMoved {
			if b.PieceAt(Position{Row: row, Col: 1}) == nil &&
				b.PieceAt(Position{Row: row, Col: 2}) == nil &&
				b.PieceAt(Position{Row: row, Col: 3}) == nil &&
				!b.IsSquareAttacked(Position{Row: row, Col: 2}, p.Color.Opponent()) &&
				!b.IsSquareAttacked(Position{Row: row, Col: 3}, p.Color.Opponent()) {
				moves = append(moves, Position{Row: row, Col: 2})
			}
		}
	}

	return moves
}

func (p *Piece) pawnMoves(b *Board) []Position {
	var moves []Position
	dir := 1
	startRow := 1
	if p.Color == Black {
		dir = -1
		startRow = 6
	}

	// Forward one square
	oneAhead := Position{Row: p.Position.Row + dir, Col: p.Position.Col}
	if oneAhead.IsValid() && b.PieceAt(oneAhead) == nil {
		moves = append(moves, oneAhead)
		// Forward two squares from starting position
		if p.Position.Row == startRow {
			twoAhead := Position{Row: p.Position.Row + 2*dir, Col: p.Position.Col}
			if b.PieceAt(twoAhead) == nil {
				moves = append(moves, twoAhead)
			}
		}
	}

	// Diagonal captures
	for _, dc := range []int{-1, 1} {
		diag := Position{Row: p.Position.Row + dir, Col: p.Position.Col + dc}
		if diag.IsValid() {
			target := b.PieceAt(diag)
			if target != nil && target.Color != p.Color {
				moves = append(moves, diag)
			}
			// En passant
			if b.EnPassantTarget != nil && *b.EnPassantTarget == diag {
				moves = append(moves, diag)
			}
		}
	}

	return moves
}
