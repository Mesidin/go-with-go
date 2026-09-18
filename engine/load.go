package engine

import (
	"fmt"
	"strings"
)

// Load builds a game from an ASCII grid. Rows are top to bottom.
// '.' empty, 'X'/'B' Black, 'O'/'W' White.
func Load(size int, komi float64, toPlay Color, lines ...string) (*Game, error) {
	g, err := New(size, komi)
	if err != nil {
		return nil, err
	}
	if len(lines) != size {
		return nil, fmt.Errorf("got %d rows, want %d", len(lines), size)
	}
	for y, line := range lines {
		line = strings.TrimSpace(line)
		if len([]rune(line)) != size {
			return nil, fmt.Errorf("row %d: got %d cols, want %d", y, len([]rune(line)), size)
		}
		for x, r := range line {
			p := Point{X: x, Y: y}
			switch r {
			case 'X', 'B':
				g.board[g.idx(p)] = Black
			case 'O', 'W':
				g.board[g.idx(p)] = White
			case '.', '+':
				g.board[g.idx(p)] = Empty
			default:
				return nil, fmt.Errorf("row %d col %d: bad char %q", y, x, r)
			}
		}
	}
	if toPlay != Black && toPlay != White {
		return nil, fmt.Errorf("toPlay must be Black or White")
	}
	g.toPlay = toPlay
	return g, nil
}
