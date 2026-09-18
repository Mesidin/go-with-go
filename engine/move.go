package engine

import "fmt"

// Move is one recorded ply: a placement, a pass, or a resignation.
type Move struct {
	Color    Color
	Point    *Point
	Pass     bool
	Resign   bool
	Captured int
	Comment  string
}

func (m Move) Coord(size int) string {
	switch {
	case m.Resign:
		return "resign"
	case m.Pass:
		return "pass"
	case m.Point != nil:
		return FormatPoint(*m.Point, size)
	default:
		return ""
	}
}

func (m Move) Label(size int) string {
	s := m.Color.String() + " " + m.Coord(size)
	if m.Captured > 0 {
		s += fmt.Sprintf(" · %s", commentPlay(m.Captured))
	} else if m.Comment != "" && !m.Pass && !m.Resign {
		s += " · " + m.Comment
	}
	return s
}

func commentPlay(captured int) string {
	if captured == 1 {
		return "captures 1"
	}
	if captured > 1 {
		return fmt.Sprintf("captures %d", captured)
	}
	return ""
}

// FormatPoint returns a standard Go coordinate (A–T skipping I, rank from the bottom).
// If size is 0, it prints (x,y) instead of a rank.
func FormatPoint(p Point, size int) string {
	files := "ABCDEFGHJKLMNOPQRST"
	if p.X < 0 || p.X >= len(files) {
		return p.String()
	}
	if size <= 0 {
		return string(files[p.X]) + fmt.Sprintf("%d", p.Y)
	}
	return string(files[p.X]) + fmt.Sprintf("%d", size-p.Y)
}
