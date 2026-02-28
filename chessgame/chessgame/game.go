package chessgame

import (
	"errors"
	"fmt"
	"strings"
)

// Move records a move in the game history
type Move struct {
	From     Position
	To       Position
	Piece    PieceType
	Captured *PieceType
	IsCheck  bool
}

func (m Move) String() string {
	s := fmt.Sprintf("%s %s→%s", m.Piece, m.From, m.To)
	if m.Captured != nil {
		s += fmt.Sprintf(" x%s", *m.Captured)
	}
	if m.IsCheck {
		s += "+"
	}
	return s
}

// Game is the main game orchestrator
type Game struct {
	Board       *Board
	Players     [2]*Player
	CurrentTurn Color
	Status      GameStatus
	MoveHistory []Move
}

func NewGame(player1Name, player2Name string) *Game {
	return &Game{
		Board:       NewBoard(),
		Players:     [2]*Player{NewPlayer(player1Name, White), NewPlayer(player2Name, Black)},
		CurrentTurn: White,
		Status:      InProgress,
	}
}

// CurrentPlayer returns the player whose turn it is
func (g *Game) CurrentPlayer() *Player {
	if g.CurrentTurn == White {
		return g.Players[0]
	}
	return g.Players[1]
}

// MakeMove attempts to make a move. from/to use algebraic notation e.g. "e2", "e4"
func (g *Game) MakeMove(fromStr, toStr string) error {
	if g.Status == Checkmate || g.Status == Stalemate {
		return errors.New("game is over")
	}

	from, err := parsePosition(fromStr)
	if err != nil {
		return err
	}
	to, err := parsePosition(toStr)
	if err != nil {
		return err
	}

	piece := g.Board.PieceAt(from)
	if piece == nil {
		return fmt.Errorf("no piece at %s", fromStr)
	}
	if piece.Color != g.CurrentTurn {
		return fmt.Errorf("it's %s's turn", g.CurrentTurn)
	}
	if !g.Board.IsLegalMove(from, to) {
		return fmt.Errorf("illegal move: %s to %s", fromStr, toStr)
	}

	// Record the move before executing
	move := Move{From: from, To: to, Piece: piece.Type}
	captured := g.Board.ApplyMove(from, to)
	if captured != nil {
		ct := captured.Type
		move.Captured = &ct
	}

	// Update game status
	opponent := g.CurrentTurn.Opponent()
	inCheck := g.Board.IsInCheck(opponent)
	hasLegalMoves := g.Board.HasAnyLegalMoves(opponent)

	if inCheck {
		move.IsCheck = true
		if !hasLegalMoves {
			g.Status = Checkmate
		} else {
			g.Status = Check
		}
	} else if !hasLegalMoves {
		g.Status = Stalemate
	} else {
		g.Status = InProgress
	}

	g.MoveHistory = append(g.MoveHistory, move)
	g.CurrentTurn = opponent
	return nil
}

// PrintStatus prints a human-readable game status
func (g *Game) PrintStatus() {
	switch g.Status {
	case Check:
		fmt.Printf("⚠️  %s is in CHECK!\n", g.CurrentTurn)
	case Checkmate:
		winner := g.CurrentTurn.Opponent()
		winnerPlayer := g.Players[0]
		if winner == Black {
			winnerPlayer = g.Players[1]
		}
		fmt.Printf("🏆 CHECKMATE! %s wins!\n", winnerPlayer.Name)
	case Stalemate:
		fmt.Println("🤝 STALEMATE! It's a draw.")
	case InProgress:
		fmt.Printf("▶️  %s's turn (%s)\n", g.CurrentPlayer().Name, g.CurrentTurn)
	}
}

// PrintHistory prints the move history
func (g *Game) PrintHistory() {
	fmt.Println("\n=== Move History ===")
	for i, m := range g.MoveHistory {
		color := "White"
		if i%2 == 1 {
			color = "Black"
		}
		fmt.Printf("%d. [%s] %s\n", i+1, color, m)
	}
}

// IsOver returns true if the game has ended
func (g *Game) IsOver() bool {
	return g.Status == Checkmate || g.Status == Stalemate
}

// ParsePosition converts algebraic notation "e4" to a Position
func ParsePosition(s string) (Position, error) {
	return parsePosition(s)
}

// parsePosition converts algebraic notation "e4" to a Position
func parsePosition(s string) (Position, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) != 2 {
		return Position{}, fmt.Errorf("invalid position: %s", s)
	}
	col := int(s[0] - 'a')
	row := int(s[1] - '1')
	pos := Position{Row: row, Col: col}
	if !pos.IsValid() {
		return Position{}, fmt.Errorf("position out of bounds: %s", s)
	}
	return pos, nil
}
