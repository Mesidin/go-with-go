package engine

import "fmt"

// Result is a Japanese (territory) score.
//
// Black = black territory + white stones Black captured (including marked dead).
// White = white territory + black stones White captured + komi.
// Stones remaining on the board do not count.
type Result struct {
	BlackTerritory int
	WhiteTerritory int
	BlackCaptures  int
	WhiteCaptures  int
	Komi           float64
	Black          float64
	White          float64
	Winner         Color // Empty means jigo (draw); with 6.5 komi this is rare
	Resigned       Color // the player who resigned, else Empty
}

// Margin is the absolute point difference. Zero on jigo or resignation.
func (r Result) Margin() float64 {
	if r.Resigned != Empty {
		return 0
	}
	if r.Black > r.White {
		return r.Black - r.White
	}
	return r.White - r.Black
}

func (r Result) String() string {
	if r.Resigned != Empty {
		return r.Resigned.Opponent().String() + " wins by resignation"
	}
	switch r.Winner {
	case Black:
		return fmt.Sprintf("Black wins by %.1f", r.Margin())
	case White:
		return fmt.Sprintf("White wins by %.1f", r.Margin())
	default:
		return "Jigo"
	}
}

// Score computes Japanese territory scoring.
// dead lists stones to treat as captured: they are added to the opponent's
// prisoners and their intersections become empty for territory.
func (g *Game) Score(dead map[Point]struct{}) Result {
	if g.resigned != Empty {
		winner := g.resigned.Opponent()
		return Result{
			Komi:     g.komi,
			Resigned: g.resigned,
			Winner:   winner,
		}
	}

	board := append([]Color(nil), g.board...)
	blackCap := g.captured[Black]
	whiteCap := g.captured[White]
	if dead != nil {
		for p := range dead {
			if !g.inBounds(p) {
				continue
			}
			switch board[g.idx(p)] {
			case Black:
				board[g.idx(p)] = Empty
				whiteCap++
			case White:
				board[g.idx(p)] = Empty
				blackCap++
			}
		}
	}

	blackTerr, whiteTerr := territory(g.size, board)

	r := Result{
		BlackTerritory: blackTerr,
		WhiteTerritory: whiteTerr,
		BlackCaptures:  blackCap,
		WhiteCaptures:  whiteCap,
		Komi:           g.komi,
		Black:          float64(blackTerr + blackCap),
		White:          float64(whiteTerr+whiteCap) + g.komi,
	}
	switch {
	case r.Black > r.White:
		r.Winner = Black
	case r.White > r.Black:
		r.Winner = White
	default:
		r.Winner = Empty
	}
	return r
}

func territory(size int, board []Color) (black, white int) {
	seen := make([]bool, len(board))
	idx := func(p Point) int { return p.Y*size + p.X }
	in := func(p Point) bool { return p.X >= 0 && p.Y >= 0 && p.X < size && p.Y < size }
	neigh := func(p Point) []Point {
		out := make([]Point, 0, 4)
		for _, d := range [4]Point{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
			n := Point{p.X + d.X, p.Y + d.Y}
			if in(n) {
				out = append(out, n)
			}
		}
		return out
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			p := Point{x, y}
			i := idx(p)
			if board[i] != Empty || seen[i] {
				continue
			}
			region := []Point{p}
			seen[i] = true
			touchesB, touchesW := false, false
			for q := 0; q < len(region); q++ {
				cur := region[q]
				for _, n := range neigh(cur) {
					switch board[idx(n)] {
					case Black:
						touchesB = true
					case White:
						touchesW = true
					case Empty:
						ni := idx(n)
						if !seen[ni] {
							seen[ni] = true
							region = append(region, n)
						}
					}
				}
			}
			switch {
			case touchesB && !touchesW:
				black += len(region)
			case touchesW && !touchesB:
				white += len(region)
			}
		}
	}
	return black, white
}
