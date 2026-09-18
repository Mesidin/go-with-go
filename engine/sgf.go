package engine

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// SGF returns a FF[4] Smart Game Format record of this game.
func (g *Game) SGF() string {
	var b strings.Builder
	b.WriteString("(;FF[4]GM[1]CA[UTF-8]")
	fmt.Fprintf(&b, "SZ[%d]KM[%g]", g.size, g.komi)
	if g.handicap >= 2 {
		fmt.Fprintf(&b, "HA[%d]AB", g.handicap)
		if pts, err := HandicapPoints(g.size, g.handicap); err == nil {
			for _, p := range pts {
				fmt.Fprintf(&b, "[%s]", sgfCoord(p))
			}
		}
	}
	b.WriteString("RU[Japanese]")
	if g.over {
		r := g.Score(nil)
		b.WriteString("RE[" + sgfResult(r) + "]")
	}
	for _, m := range g.moves {
		tag := "B"
		if m.Color == White {
			tag = "W"
		}
		switch {
		case m.Resign:
			// Resignation is stored in RE; skip a ply.
		case m.Pass:
			fmt.Fprintf(&b, ";%s[]", tag)
		case m.Point != nil:
			fmt.Fprintf(&b, ";%s[%s]", tag, sgfCoord(*m.Point))
		}
	}
	b.WriteByte(')')
	return b.String()
}

func sgfResult(r Result) string {
	if r.Resigned != Empty {
		if r.Resigned == Black {
			return "W+R"
		}
		return "B+R"
	}
	switch r.Winner {
	case Black:
		return fmt.Sprintf("B+%.1f", r.Margin())
	case White:
		return fmt.Sprintf("W+%.1f", r.Margin())
	default:
		return "0"
	}
}

func sgfCoord(p Point) string {
	return string(rune('a'+p.X)) + string(rune('a'+p.Y))
}

func parseSGFCoord(s string, size int) (Point, bool, error) {
	if s == "" || s == "tt" && size <= 19 {
		return Point{}, true, nil // pass
	}
	if len(s) != 2 {
		return Point{}, false, fmt.Errorf("bad SGF coord %q", s)
	}
	x := int(s[0] - 'a')
	y := int(s[1] - 'a')
	if x < 0 || y < 0 || x >= size || y >= size {
		return Point{}, false, fmt.Errorf("SGF coord %q off board", s)
	}
	return Point{X: x, Y: y}, false, nil
}

// ParseSGF reads a single-game FF[4] record (main line only).
func ParseSGF(data string) (*Game, error) {
	p := sgfParser{s: data}
	props, moves, err := p.game()
	if err != nil {
		return nil, err
	}
	size := 19
	if v, ok := props["SZ"]; ok && len(v) > 0 {
		size, err = strconv.Atoi(v[0])
		if err != nil {
			return nil, fmt.Errorf("SZ: %w", err)
		}
	}
	komi := DefaultKomi
	if v, ok := props["KM"]; ok && len(v) > 0 {
		komi, err = strconv.ParseFloat(v[0], 64)
		if err != nil {
			return nil, fmt.Errorf("KM: %w", err)
		}
	}
	g, err := New(size, komi)
	if err != nil {
		return nil, err
	}
	if v, ok := props["HA"]; ok && len(v) > 0 {
		ha, err := strconv.Atoi(v[0])
		if err != nil {
			return nil, fmt.Errorf("HA: %w", err)
		}
		g.handicap = ha
	}
	if v := props["AB"]; len(v) > 0 {
		for _, c := range v {
			pt, pass, err := parseSGFCoord(c, size)
			if err != nil || pass {
				return nil, fmt.Errorf("AB: %w", err)
			}
			g.board[g.idx(pt)] = Black
		}
		if g.handicap >= 2 {
			g.toPlay = White
		}
	} else if g.handicap >= 2 {
		if err := g.PlaceHandicap(g.handicap); err != nil {
			return nil, err
		}
	}
	if v := props["AW"]; len(v) > 0 {
		for _, c := range v {
			pt, pass, err := parseSGFCoord(c, size)
			if err != nil || pass {
				return nil, fmt.Errorf("AW: %w", err)
			}
			g.board[g.idx(pt)] = White
		}
	}
	if v, ok := props["PL"]; ok && len(v) > 0 {
		switch strings.ToUpper(v[0]) {
		case "B", "BLACK":
			g.toPlay = Black
		case "W", "WHITE":
			g.toPlay = White
		}
	}
	for _, m := range moves {
		switch {
		case m.resign:
			_ = g.Resign()
		case m.pass:
			if err := g.Pass(); err != nil {
				return nil, err
			}
		default:
			if err := g.Play(m.pt); err != nil {
				return nil, fmt.Errorf("move %s: %w", FormatPoint(m.pt, size), err)
			}
		}
	}
	return g, nil
}

type sgfMove struct {
	pt     Point
	pass   bool
	resign bool
}

type sgfParser struct {
	s string
	i int
}

func (p *sgfParser) game() (map[string][]string, []sgfMove, error) {
	p.skip()
	if !p.eat('(') {
		return nil, nil, fmt.Errorf("SGF: expected '('")
	}
	p.skip()
	if !p.eat(';') {
		return nil, nil, fmt.Errorf("SGF: expected root ';'")
	}
	root, err := p.props()
	if err != nil {
		return nil, nil, err
	}
	var moves []sgfMove
	for {
		p.skip()
		if p.eat(';') {
			pr, err := p.props()
			if err != nil {
				return nil, nil, err
			}
			if v := pr["B"]; len(v) > 0 {
				pt, pass, err := parseSGFCoord(v[0], 19)
				if err != nil {
					// size-checked later in Play; accept a-s
					if len(v[0]) == 2 {
						pt = Point{X: int(v[0][0] - 'a'), Y: int(v[0][1] - 'a')}
						pass = false
						err = nil
					}
				}
				if err != nil {
					return nil, nil, err
				}
				moves = append(moves, sgfMove{pt: pt, pass: pass})
			} else if v := pr["W"]; len(v) > 0 {
				pt, pass, err := parseSGFCoord(v[0], 19)
				if err != nil && len(v[0]) == 2 {
					pt = Point{X: int(v[0][0] - 'a'), Y: int(v[0][1] - 'a')}
					pass, err = false, nil
				}
				if err != nil {
					return nil, nil, err
				}
				moves = append(moves, sgfMove{pt: pt, pass: pass})
			}
			continue
		}
		if p.eat('(') {
			// skip variation: consume until matching )
			depth := 1
			for depth > 0 && p.i < len(p.s) {
				switch p.s[p.i] {
				case '(':
					depth++
				case ')':
					depth--
				case '[':
					p.i++
					for p.i < len(p.s) && p.s[p.i] != ']' {
						if p.s[p.i] == '\\' {
							p.i++
						}
						p.i++
					}
				}
				p.i++
			}
			continue
		}
		if p.eat(')') {
			break
		}
		if p.i >= len(p.s) {
			return nil, nil, fmt.Errorf("SGF: unexpected end")
		}
		return nil, nil, fmt.Errorf("SGF: unexpected %q", p.s[p.i])
	}
	return root, moves, nil
}

func (p *sgfParser) props() (map[string][]string, error) {
	out := map[string][]string{}
	for {
		p.skip()
		if p.i >= len(p.s) {
			break
		}
		c := p.s[p.i]
		if c == ';' || c == '(' || c == ')' {
			break
		}
		ident := p.ident()
		if ident == "" {
			return nil, fmt.Errorf("SGF: expected property at %d", p.i)
		}
		var vals []string
		for {
			p.skip()
			if !p.eat('[') {
				break
			}
			v, err := p.value()
			if err != nil {
				return nil, err
			}
			vals = append(vals, v)
		}
		if len(vals) == 0 {
			return nil, fmt.Errorf("SGF: property %s has no value", ident)
		}
		out[ident] = append(out[ident], vals...)
	}
	return out, nil
}

func (p *sgfParser) ident() string {
	start := p.i
	for p.i < len(p.s) && p.s[p.i] >= 'A' && p.s[p.i] <= 'Z' {
		p.i++
	}
	return p.s[start:p.i]
}

func (p *sgfParser) value() (string, error) {
	var b strings.Builder
	for p.i < len(p.s) {
		c := p.s[p.i]
		p.i++
		switch c {
		case ']':
			return b.String(), nil
		case '\\':
			if p.i < len(p.s) {
				b.WriteByte(p.s[p.i])
				p.i++
			}
		default:
			b.WriteByte(c)
		}
	}
	return "", fmt.Errorf("SGF: unterminated value")
}

func (p *sgfParser) skip() {
	for p.i < len(p.s) {
		c := p.s[p.i]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' || unicode.IsSpace(rune(c)) {
			p.i++
			continue
		}
		break
	}
}

func (p *sgfParser) eat(c byte) bool {
	if p.i < len(p.s) && p.s[p.i] == c {
		p.i++
		return true
	}
	return false
}
