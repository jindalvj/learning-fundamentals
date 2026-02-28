package chessgame

import (
	"testing"
)

func TestNewBoard(t *testing.T) {
	b := NewBoard()

	// Check white king is at e1
	king := b.PieceAt(Position{0, 4})
	if king == nil || king.Type != King || king.Color != White {
		t.Error("expected white king at e1")
	}

	// Check black king is at e8
	bking := b.PieceAt(Position{7, 4})
	if bking == nil || bking.Type != King || bking.Color != Black {
		t.Error("expected black king at e8")
	}

	// Check pawn at e2
	pawn := b.PieceAt(Position{1, 4})
	if pawn == nil || pawn.Type != Pawn || pawn.Color != White {
		t.Error("expected white pawn at e2")
	}
}

func TestPawnMove(t *testing.T) {
	b := NewBoard()
	from := Position{1, 4} // e2
	to := Position{3, 4}   // e4

	if !b.IsLegalMove(from, to) {
		t.Error("double pawn push from e2 to e4 should be legal")
	}

	b.ApplyMove(from, to)
	if b.PieceAt(to) == nil {
		t.Error("pawn should be at e4")
	}
	if b.PieceAt(from) != nil {
		t.Error("e2 should be empty")
	}
}

func TestIllegalMoveIntoCheck(t *testing.T) {
	// Setup a position where moving a piece exposes king
	b := &Board{}
	b.placePiece(NewPiece(King, White, Position{0, 4}))
	b.placePiece(NewPiece(Rook, White, Position{0, 3})) // shields king from black rook
	b.placePiece(NewPiece(Rook, Black, Position{0, 0})) // attacks along rank 1

	// Moving the white rook away exposes the king
	if b.IsLegalMove(Position{0, 3}, Position{4, 3}) {
		t.Error("moving shielding rook should be illegal (exposes king)")
	}
}

func TestKnightMoves(t *testing.T) {
	b := NewBoard()
	from := Position{0, 1} // b1
	moves := b.LegalMovesFor(from)

	// Knight on b1 can move to a3 or c3
	found := map[string]bool{}
	for _, m := range moves {
		found[m.String()] = true
	}
	if !found["a3"] || !found["c3"] {
		t.Errorf("knight at b1 should be able to reach a3 and c3, got %v", moves)
	}
}

func TestCheckDetection(t *testing.T) {
	b := &Board{}
	b.placePiece(NewPiece(King, White, Position{0, 4}))
	b.placePiece(NewPiece(Queen, Black, Position{0, 7})) // attacks rank 1

	if !b.IsInCheck(White) {
		t.Error("white king should be in check from black queen on same rank")
	}
}

func TestCheckmateScholars(t *testing.T) {
	// Scholar's mate: 1.e4 e5 2.Bc4 Nc6 3.Qh5 Nf6?? 4.Qxf7#
	g := NewGame("Alice", "Bob")
	moves := [][2]string{
		{"e2", "e4"}, {"e7", "e5"},
		{"f1", "c4"}, {"b8", "c6"},
		{"d1", "h5"}, {"g8", "f6"},
		{"h5", "f7"},
	}
	for _, m := range moves {
		if err := g.MakeMove(m[0], m[1]); err != nil {
			t.Fatalf("unexpected error on move %v: %v", m, err)
		}
	}
	if g.Status != Checkmate {
		t.Errorf("expected Checkmate, got %v", g.Status)
	}
}

func TestEnPassant(t *testing.T) {
	g := NewGame("A", "B")
	// 1.e4 a5 2.e5 d5 (black's d-pawn double-pushes next to white's e-pawn)
	moves := [][2]string{
		{"e2", "e4"}, {"a7", "a5"},
		{"e4", "e5"}, {"d7", "d5"},
		{"e5", "d6"}, // en passant
	}
	for _, m := range moves {
		if err := g.MakeMove(m[0], m[1]); err != nil {
			t.Fatalf("unexpected error on move %v: %v", m, err)
		}
	}
	// Black pawn on d5 should be captured
	if g.Board.PieceAt(Position{4, 3}) != nil {
		t.Error("black pawn on d5 should have been captured by en passant")
	}
	// White pawn should be on d6
	wp := g.Board.PieceAt(Position{5, 3})
	if wp == nil || wp.Color != White {
		t.Error("white pawn should be on d6 after en passant")
	}
}

func TestCastlingKingside(t *testing.T) {
	b := &Board{}
	b.placePiece(NewPiece(King, White, Position{0, 4}))
	b.placePiece(NewPiece(Rook, White, Position{0, 7}))

	if !b.IsLegalMove(Position{0, 4}, Position{0, 6}) {
		t.Error("king-side castling should be legal")
	}
	b.ApplyMove(Position{0, 4}, Position{0, 6})

	king := b.PieceAt(Position{0, 6})
	rook := b.PieceAt(Position{0, 5})
	if king == nil || king.Type != King {
		t.Error("king should be on g1")
	}
	if rook == nil || rook.Type != Rook {
		t.Error("rook should be on f1 after king-side castle")
	}
}

func TestStalemate(t *testing.T) {
	// Classic stalemate position: black king only legal square is attacked
	b := &Board{}
	b.placePiece(NewPiece(King, White, Position{5, 5})) // f6
	b.placePiece(NewPiece(Queen, White, Position{6, 1})) // b7
	b.placePiece(NewPiece(King, Black, Position{7, 0})) // a8

	// Black has no legal moves and is not in check
	if b.IsInCheck(Black) {
		t.Error("black should not be in check in stalemate position")
	}
	if b.HasAnyLegalMoves(Black) {
		t.Error("black should have no legal moves in stalemate")
	}
}
