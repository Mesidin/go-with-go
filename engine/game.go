package engine

import (
	"errors"
	"fmt"
)

var (
	ErrOffBoard    = errors.New("point is off the board")
	ErrOccupied    = errors.New("point is occupied")
	ErrKo          = errors.New("move violates the simple ko rule")
	ErrSuicide     = errors.New("move is suicide")
	ErrGameOver    = errors.New("game is already over")
	ErrNothingUndo = errors.New("nothing to undo")
	ErrBadSize     = errors.New("board size must be 9, 13, or 19")
)

// DefaultKomi is Japanese komi, used on every board size.
const DefaultKomi = 6.5

// Game is a club-rules game of Go with Japanese (territory) scoring.
type Game struct {
	size               int
	komi               float64
	board              []Color
	toPlay             Color
	ko                 *Point
	captured           [3]int // indexed by Color: stones that color has taken
	consecutivePasses  int
	over               bool
	resigned           Color // Empty if nobody resigned
	lastPoint          *Point
	history            []snapshot
	handicap           int
	moves              []Move
}

type snapshot struct {
	board             []Color
	toPlay            Color
	ko                *Point
	captured          [3]int
	consecutivePasses int
	over              bool
	resigned          Color
	lastPoint         *Point
}

// New starts an empty game. size must be 9, 13, or 19. komi is added to White.
func New(size int, komi float64) (*Game, error) {
	if !validSize(size) {
		return nil, fmt.Errorf("%w: %d", ErrBadSize, size)
	}
	g := &Game{
		size:   size,
		komi:   komi,
		board:  make([]Color, size*size),
		toPlay: Black,
	}
	return g, nil
}

func (g *Game) Size() int         { return g.size }
func (g *Game) Komi() float64     { return g.komi }
func (g *Game) ToPlay() Color     { return g.toPlay }
func (g *Game) Over() bool        { return g.over }
func (g *Game) Resigned() Color   { return g.resigned }
func (g *Game) LastPoint() *Point { return clonePoint(g.lastPoint) }
func (g *Game) Ko() *Point        { return clonePoint(g.ko) }
func (g *Game) Handicap() int     { return g.handicap }
func (g *Game) Moves() []Move     { return append([]Move(nil), g.moves...) }
func (g *Game) MoveCount() int    { return len(g.moves) }

func (g *Game) LastMove() *Move {
	if len(g.moves) == 0 {
		return nil
	}
	m := g.moves[len(g.moves)-1]
	return &m
}

// Neighbors returns the on-board 4-adjacent points.
func (g *Game) Neighbors(p Point) []Point { return g.neighbors(p) }

// Group returns the connected stones of the same color as p.
func (g *Game) Group(p Point) []Point { return g.group(p) }

// LibertyCount is the number of distinct liberties of the group containing p.
func (g *Game) LibertyCount(p Point) int { return g.libertyCount(p) }

// At returns the color at p, or Empty if p is off-board.
func (g *Game) At(p Point) Color {
	if !g.inBounds(p) {
		return Empty
	}
	return g.board[g.idx(p)]
}

// Captured returns how many stones color has taken (prisoners for that player).
func (g *Game) Captured(c Color) int {
	if c != Black && c != White {
		return 0
	}
	return g.captured[c]
}

func (g *Game) Clone() *Game {
	cp := *g
	cp.board = append([]Color(nil), g.board...)
	cp.ko = clonePoint(g.ko)
	cp.lastPoint = clonePoint(g.lastPoint)
	cp.moves = append([]Move(nil), g.moves...)
	cp.history = append([]snapshot(nil), g.history...)
	for i := range cp.history {
		cp.history[i].board = append([]Color(nil), g.history[i].board...)
		cp.history[i].ko = clonePoint(g.history[i].ko)
		cp.history[i].lastPoint = clonePoint(g.history[i].lastPoint)
	}
	return &cp
}

func clonePoint(p *Point) *Point {
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}

func (g *Game) save() {
	s := snapshot{
		board:             append([]Color(nil), g.board...),
		toPlay:            g.toPlay,
		ko:                clonePoint(g.ko),
		captured:          g.captured,
		consecutivePasses: g.consecutivePasses,
		over:              g.over,
		resigned:          g.resigned,
		lastPoint:         clonePoint(g.lastPoint),
	}
	g.history = append(g.history, s)
}

// Undo reverts the last ply (a play, pass, or resign).
func (g *Game) Undo() error {
	if len(g.history) == 0 {
		return ErrNothingUndo
	}
	s := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.board = append([]Color(nil), s.board...)
	g.toPlay = s.toPlay
	g.ko = clonePoint(s.ko)
	g.captured = s.captured
	g.consecutivePasses = s.consecutivePasses
	g.over = s.over
	g.resigned = s.resigned
	g.lastPoint = clonePoint(s.lastPoint)
	if len(g.moves) > 0 {
		g.moves = g.moves[:len(g.moves)-1]
	}
	return nil
}

// Legal reports whether a placement at p is allowed for the player to move.
func (g *Game) Legal(p Point) bool {
	return g.legalErr(p) == nil
}

func (g *Game) legalErr(p Point) error {
	if g.over {
		return ErrGameOver
	}
	if !g.inBounds(p) {
		return ErrOffBoard
	}
	if g.board[g.idx(p)] != Empty {
		return ErrOccupied
	}
	if g.ko != nil && *g.ko == p {
		return ErrKo
	}
	if g.wouldSuicide(p) {
		return ErrSuicide
	}
	return nil
}

func (g *Game) wouldSuicide(p Point) bool {
	trial := g.Clone()
	trial.board[trial.idx(p)] = g.toPlay
	trial.removeDead(g.toPlay.Opponent())
	return trial.libertyCount(p) == 0
}

// Play places a stone for the player to move.
func (g *Game) Play(p Point) error {
	if err := g.legalErr(p); err != nil {
		return err
	}
	g.save()
	who := g.toPlay
	g.board[g.idx(p)] = who
	n := g.removeDead(who.Opponent())
	g.captured[who] += n
	g.ko = g.simpleKo(p, n)
	g.consecutivePasses = 0
	g.lastPoint = &Point{X: p.X, Y: p.Y}
	g.moves = append(g.moves, Move{
		Color:    who,
		Point:    &Point{X: p.X, Y: p.Y},
		Captured: n,
		Comment:  commentPlay(n),
	})
	g.toPlay = g.toPlay.Opponent()
	return nil
}

func (g *Game) removeDead(c Color) int {
	seen := make([]bool, len(g.board))
	removed := 0
	for y := 0; y < g.size; y++ {
		for x := 0; x < g.size; x++ {
			p := Point{x, y}
			i := g.idx(p)
			if g.board[i] != c || seen[i] {
				continue
			}
			grp := g.group(p)
			for _, q := range grp {
				seen[g.idx(q)] = true
			}
			if g.libertyCount(p) > 0 {
				continue
			}
			for _, q := range grp {
				g.board[g.idx(q)] = Empty
				removed++
			}
		}
	}
	return removed
}

func (g *Game) simpleKo(played Point, captured int) *Point {
	if captured != 1 {
		return nil
	}
	if len(g.group(played)) != 1 {
		return nil
	}
	// The single captured stone is the only empty neighbor of the placed stone
	// that was opponent-colored before the capture — equivalently, the unique
	// empty neighbor after capture, for a size-1 group that took one stone.
	empties := g.liberties(played)
	if len(empties) != 1 {
		return nil
	}
	return &empties[0]
}

// Pass is always legal while the game is ongoing. Two consecutive passes end it.
func (g *Game) Pass() error {
	if g.over {
		return ErrGameOver
	}
	g.save()
	who := g.toPlay
	g.ko = nil
	g.lastPoint = nil
	g.consecutivePasses++
	g.moves = append(g.moves, Move{Color: who, Pass: true, Comment: "pass"})
	g.toPlay = who.Opponent()
	if g.consecutivePasses >= 2 {
		g.over = true
	}
	return nil
}

// Resign ends the game; the opponent wins.
func (g *Game) Resign() error {
	if g.over {
		return ErrGameOver
	}
	g.save()
	who := g.toPlay
	g.resigned = who
	g.over = true
	g.lastPoint = nil
	g.moves = append(g.moves, Move{Color: who, Resign: true, Comment: "resign"})
	return nil
}

// LegalMoves returns every legal placement for the player to move.
func (g *Game) LegalMoves() []Point {
	if g.over {
		return nil
	}
	var out []Point
	for y := 0; y < g.size; y++ {
		for x := 0; x < g.size; x++ {
			p := Point{x, y}
			if g.Legal(p) {
				out = append(out, p)
			}
		}
	}
	return out
}

// ResumeAfterPasses undoes a two-pass ending so play can continue.
// Used when players passed into scoring and then chose to resume.
func (g *Game) ResumeAfterPasses() error {
	if !g.over || g.resigned != Empty || g.consecutivePasses < 2 {
		return errors.New("game is not in a two-pass ending")
	}
	g.over = false
	g.consecutivePasses = 0
	return nil
}
