package engine

import "fmt"

var (
	ErrBadHandicap = fmt.Errorf("handicap must be 0 or 2–9")
)

// PlaceHandicap puts Black stones on the star points and, for 2 or more,
// makes White play first. It does not record moves (SGF uses AB setup).
func (g *Game) PlaceHandicap(n int) error {
	if n == 0 {
		g.handicap = 0
		g.toPlay = Black
		return nil
	}
	pts, err := HandicapPoints(g.size, n)
	if err != nil {
		return err
	}
	for _, p := range pts {
		if g.board[g.idx(p)] != Empty {
			return fmt.Errorf("handicap point %s is occupied", p)
		}
		g.board[g.idx(p)] = Black
	}
	g.handicap = n
	g.toPlay = White
	return nil
}

// HandicapPoints returns standard Japanese star-point placements.
func HandicapPoints(size, n int) ([]Point, error) {
	if n == 1 || n < 0 || n > 9 {
		return nil, fmt.Errorf("%w: %d", ErrBadHandicap, n)
	}
	if n == 0 {
		return nil, nil
	}
	var d, t int // corner offset and tengen
	switch size {
	case 9:
		d, t = 2, 4
	case 13:
		d, t = 3, 6
	case 19:
		d, t = 3, 9
	default:
		return nil, fmt.Errorf("%w: %d", ErrBadSize, size)
	}
	hi := size - 1 - d
	// Order: lower-left, upper-right, upper-left, lower-right, tengen,
	// left, right, bottom, top (from Black's side, y=0 at top).
	ll := Point{X: d, Y: hi}
	ur := Point{X: hi, Y: d}
	ul := Point{X: d, Y: d}
	lr := Point{X: hi, Y: hi}
	c := Point{X: t, Y: t}
	left := Point{X: d, Y: t}
	right := Point{X: hi, Y: t}
	bot := Point{X: t, Y: hi}
	top := Point{X: t, Y: d}

	var out []Point
	switch n {
	case 2:
		out = []Point{ll, ur}
	case 3:
		out = []Point{ll, ur, ul}
	case 4:
		out = []Point{ll, ur, ul, lr}
	case 5:
		out = []Point{ll, ur, ul, lr, c}
	case 6:
		out = []Point{ll, ur, ul, lr, left, right}
	case 7:
		out = []Point{ll, ur, ul, lr, left, right, c}
	case 8:
		out = []Point{ll, ur, ul, lr, left, right, bot, top}
	case 9:
		out = []Point{ll, ur, ul, lr, left, right, bot, top, c}
	}
	return out, nil
}
