package chessgame

// Player represents a chess player
type Player struct {
	Name  string
	Color Color
}

func NewPlayer(name string, color Color) *Player {
	return &Player{Name: name, Color: color}
}
