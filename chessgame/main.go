package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"chessgame/chessgame"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("♟  Welcome to Chess!")
	fmt.Println("═══════════════════")

	fmt.Print("Player 1 (White) name: ")
	p1, _ := reader.ReadString('\n')
	p1 = strings.TrimSpace(p1)
	if p1 == "" {
		p1 = "Player 1"
	}

	fmt.Print("Player 2 (Black) name: ")
	p2, _ := reader.ReadString('\n')
	p2 = strings.TrimSpace(p2)
	if p2 == "" {
		p2 = "Player 2"
	}

	game := chessgame.NewGame(p1, p2)

	fmt.Println("\nCommands:")
	fmt.Println("  <from> <to>  — move a piece, e.g. 'e2 e4'")
	fmt.Println("  moves <sq>   — show legal moves for a square, e.g. 'moves e2'")
	fmt.Println("  history      — show move history")
	fmt.Println("  quit         — exit")
	fmt.Println()

	for !game.IsOver() {
		game.Board.Display()
		game.PrintStatus()

		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return

		case "history":
			game.PrintHistory()

		case "moves":
			if len(parts) < 2 {
				fmt.Println("Usage: moves <square>  e.g. moves e2")
				continue
			}
			pos, err := chessgame.ParsePosition(parts[1])
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			legal := game.Board.LegalMovesFor(pos)
			if len(legal) == 0 {
				fmt.Println("No legal moves for", parts[1])
			} else {
				strs := make([]string, len(legal))
				for i, p := range legal {
					strs[i] = p.String()
				}
				fmt.Printf("Legal moves: %s\n", strings.Join(strs, ", "))
			}

		default:
			if len(parts) == 2 {
				err := game.MakeMove(parts[0], parts[1])
				if err != nil {
					fmt.Println("❌", err)
				}
			} else {
				fmt.Println("Unknown command. Enter moves like 'e2 e4' or type 'help'")
			}
		}
	}

	// Final board state
	game.Board.Display()
	game.PrintStatus()
	game.PrintHistory()
}
