package chessgame

import "fmt"

// Board represents the 8x8 chess board
type Board struct {
	squares         [8][8]*Piece
	EnPassantTarget *Position // square where en passant capture is possible
}

func NewBoard() *Board {
	b := &Board{}
	b.setup()
	return b
}

func (b *Board) setup() {
	// White pieces (row 0 = rank 1)
	b.placePiece(NewPiece(Rook, White, Position{0, 0}))
	b.placePiece(NewPiece(Knight, White, Position{0, 1}))
	b.placePiece(NewPiece(Bishop, White, Position{0, 2}))
	b.placePiece(NewPiece(Queen, White, Position{0, 3}))
	b.placePiece(NewPiece(King, White, Position{0, 4}))
	b.placePiece(NewPiece(Bishop, White, Position{0, 5}))
	b.placePiece(NewPiece(Knight, White, Position{0, 6}))
	b.placePiece(NewPiece(Rook, White, Position{0, 7}))
	for c := 0; c < 8; c++ {
		b.placePiece(NewPiece(Pawn, White, Position{1, c}))
	}

	// Black pieces (row 7 = rank 8)
	b.placePiece(NewPiece(Rook, Black, Position{7, 0}))
	b.placePiece(NewPiece(Knight, Black, Position{7, 1}))
	b.placePiece(NewPiece(Bishop, Black, Position{7, 2}))
	b.placePiece(NewPiece(Queen, Black, Position{7, 3}))
	b.placePiece(NewPiece(King, Black, Position{7, 4}))
	b.placePiece(NewPiece(Bishop, Black, Position{7, 5}))
	b.placePiece(NewPiece(Knight, Black, Position{7, 6}))
	b.placePiece(NewPiece(Rook, Black, Position{7, 7}))
	for c := 0; c < 8; c++ {
		b.placePiece(NewPiece(Pawn, Black, Position{6, c}))
	}
}

func (b *Board) placePiece(p *Piece) {
	b.squares[p.Position.Row][p.Position.Col] = p
}

// PieceAt returns the piece at the given position, or nil
func (b *Board) PieceAt(pos Position) *Piece {
	if !pos.IsValid() {
		return nil
	}
	return b.squares[pos.Row][pos.Col]
}

// Clone creates a deep copy of the board for simulating moves
func (b *Board) Clone() *Board {
	nb := &Board{}
	if b.EnPassantTarget != nil {
		ep := *b.EnPassantTarget
		nb.EnPassantTarget = &ep
	}
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if b.squares[r][c] != nil {
				p := *b.squares[r][c]
				nb.squares[r][c] = &p
			}
		}
	}
	return nb
}

// ApplyMove executes a move on the board (no legality check here — call IsLegalMove first)
// Returns the captured piece if any
func (b *Board) ApplyMove(from, to Position) *Piece {
	piece := b.squares[from.Row][from.Col]
	if piece == nil {
		return nil
	}

	b.EnPassantTarget = nil
	captured := b.squares[to.Row][to.Col]

	// En passant capture
	if piece.Type == Pawn && captured == nil && from.Col != to.Col {
		epRow := from.Row // the captured pawn is on the same row as attacker
		captured = b.squares[epRow][to.Col]
		b.squares[epRow][to.Col] = nil
	}

	// Double pawn push — set en passant target
	if piece.Type == Pawn && abs(to.Row-from.Row) == 2 {
		epRow := (from.Row + to.Row) / 2
		ep := Position{Row: epRow, Col: from.Col}
		b.EnPassantTarget = &ep
	}

	// Castling — move the rook too
	if piece.Type == King && abs(to.Col-from.Col) == 2 {
		if to.Col == 6 { // king-side
			rook := b.squares[from.Row][7]
			b.squares[from.Row][5] = rook
			b.squares[from.Row][7] = nil
			if rook != nil {
				rook.Position = Position{Row: from.Row, Col: 5}
				rook.HasMoved = true
			}
		} else { // queen-side
			rook := b.squares[from.Row][0]
			b.squares[from.Row][3] = rook
			b.squares[from.Row][0] = nil
			if rook != nil {
				rook.Position = Position{Row: from.Row, Col: 3}
				rook.HasMoved = true
			}
		}
	}

	b.squares[to.Row][to.Col] = piece
	b.squares[from.Row][from.Col] = nil
	piece.Position = to
	piece.HasMoved = true

	// Pawn promotion — auto-promote to Queen
	if piece.Type == Pawn && (to.Row == 0 || to.Row == 7) {
		piece.Type = Queen
	}

	return captured
}

// IsSquareAttacked returns true if the square is attacked by the given attacker color
func (b *Board) IsSquareAttacked(pos Position, byColor Color) bool {
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := b.squares[r][c]
			if p == nil || p.Color != byColor {
				continue
			}
			for _, m := range p.PotentialMoves(b) {
				if m == pos {
					return true
				}
			}
		}
	}
	return false
}

// IsInCheck returns true if the given color's king is currently in check
func (b *Board) IsInCheck(color Color) bool {
	kingPos := b.findKing(color)
	if kingPos == nil {
		return false
	}
	return b.IsSquareAttacked(*kingPos, color.Opponent())
}

func (b *Board) findKing(color Color) *Position {
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := b.squares[r][c]
			if p != nil && p.Type == King && p.Color == color {
				pos := Position{r, c}
				return &pos
			}
		}
	}
	return nil
}

// IsLegalMove checks if a move is legal (doesn't leave own king in check)
func (b *Board) IsLegalMove(from, to Position) bool {
	piece := b.PieceAt(from)
	if piece == nil {
		return false
	}
	// Check that 'to' is in the piece's potential moves
	found := false
	for _, m := range piece.PotentialMoves(b) {
		if m == to {
			found = true
			break
		}
	}
	if !found {
		return false
	}
	// Simulate move and check for self-check
	clone := b.Clone()
	clone.ApplyMove(from, to)
	return !clone.IsInCheck(piece.Color)
}

// LegalMovesFor returns all legal moves for a piece
func (b *Board) LegalMovesFor(from Position) []Position {
	piece := b.PieceAt(from)
	if piece == nil {
		return nil
	}
	var legal []Position
	for _, to := range piece.PotentialMoves(b) {
		if b.IsLegalMove(from, to) {
			legal = append(legal, to)
		}
	}
	return legal
}

// HasAnyLegalMoves returns true if the given color has at least one legal move
func (b *Board) HasAnyLegalMoves(color Color) bool {
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := b.squares[r][c]
			if p != nil && p.Color == color {
				if len(b.LegalMovesFor(Position{r, c})) > 0 {
					return true
				}
			}
		}
	}
	return false
}

// Display prints the board to stdout
func (b *Board) Display() {
	fmt.Println()
	fmt.Println("  a b c d e f g h")
	fmt.Println("  +-+-+-+-+-+-+-+")
	for r := 7; r >= 0; r-- {
		fmt.Printf("%d|", r+1)
		for c := 0; c < 8; c++ {
			p := b.squares[r][c]
			if p == nil {
				if (r+c)%2 == 0 {
					fmt.Print(".")
				} else {
					fmt.Print(" ")
				}
			} else {
				fmt.Print(p.Symbol())
			}
			if c < 7 {
				fmt.Print(" ")
			}
		}
		fmt.Printf("|%d\n", r+1)
	}
	fmt.Println("  +-+-+-+-+-+-+-+")
	fmt.Println("  a b c d e f g h")
	fmt.Println()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
