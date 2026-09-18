package engine

import "fmt"

// Color is the occupant of a point, or Empty.
type Color uint8

const (
	Empty Color = iota
	Black
	White
)

func (c Color) String() string {
	switch c {
	case Black:
		return "Black"
	case White:
		return "White"
	default:
		return "Empty"
	}
}

// Opponent returns the other player color. Empty is unchanged.
func (c Color) Opponent() Color {
	switch c {
	case Black:
		return White
	case White:
		return Black
	default:
		return Empty
	}
}

// Point is a board intersection. X is column, Y is row; (0,0) is top-left.
type Point struct {
	X, Y int
}

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

// ValidSizes are the supported board sizes.
var ValidSizes = []int{9, 13, 19}

func validSize(size int) bool {
	for _, s := range ValidSizes {
		if s == size {
			return true
		}
	}
	return false
}

func (g *Game) idx(p Point) int {
	return p.Y*g.size + p.X
}

func (g *Game) inBounds(p Point) bool {
	return p.X >= 0 && p.Y >= 0 && p.X < g.size && p.Y < g.size
}

func (g *Game) neighbors(p Point) []Point {
	out := make([]Point, 0, 4)
	for _, d := range [4]Point{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
		n := Point{p.X + d.X, p.Y + d.Y}
		if g.inBounds(n) {
			out = append(out, n)
		}
	}
	return out
}

// group returns every point of the same color connected to p (4-way).
func (g *Game) group(p Point) []Point {
	c := g.board[g.idx(p)]
	if c == Empty {
		return nil
	}
	seen := make([]bool, len(g.board))
	var out []Point
	var walk func(Point)
	walk = func(q Point) {
		i := g.idx(q)
		if seen[i] {
			return
		}
		if g.board[i] != c {
			return
		}
		seen[i] = true
		out = append(out, q)
		for _, n := range g.neighbors(q) {
			walk(n)
		}
	}
	walk(p)
	return out
}

// liberties returns the distinct empty points adjacent to the group containing p.
func (g *Game) liberties(p Point) []Point {
	seen := map[Point]struct{}{}
	var out []Point
	for _, q := range g.group(p) {
		for _, n := range g.neighbors(q) {
			if g.board[g.idx(n)] != Empty {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			out = append(out, n)
		}
	}
	return out
}

func (g *Game) libertyCount(p Point) int {
	return len(g.liberties(p))
}

// Hoshi returns the star-point intersections for the board size.
func Hoshi(size int) []Point {
	switch size {
	case 9:
		return []Point{{2, 2}, {6, 2}, {4, 4}, {2, 6}, {6, 6}}
	case 13:
		return []Point{{3, 3}, {9, 3}, {6, 6}, {3, 9}, {9, 9}}
	case 19:
		return []Point{
			{3, 3}, {9, 3}, {15, 3},
			{3, 9}, {9, 9}, {15, 9},
			{3, 15}, {9, 15}, {15, 15},
		}
	default:
		return nil
	}
}
