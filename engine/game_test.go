package engine

import (
	"errors"
	"strings"
	"testing"
)

func load(t *testing.T, size int, toPlay Color, lines ...string) *Game {
	t.Helper()
	g, err := Load(size, DefaultKomi, toPlay, lines...)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func grid(g *Game) string {
	var b strings.Builder
	for y := 0; y < g.size; y++ {
		if y > 0 {
			b.WriteByte('\n')
		}
		for x := 0; x < g.size; x++ {
			switch g.At(Point{x, y}) {
			case Black:
				b.WriteByte('X')
			case White:
				b.WriteByte('O')
			default:
				b.WriteByte('.')
			}
		}
	}
	return b.String()
}

func TestNewRejectsBadSize(t *testing.T) {
	if _, err := New(10, DefaultKomi); !errors.Is(err, ErrBadSize) {
		t.Fatalf("got %v, want ErrBadSize", err)
	}
}

func TestPlayAndTurn(t *testing.T) {
	g, err := New(9, DefaultKomi)
	if err != nil {
		t.Fatal(err)
	}
	if g.ToPlay() != Black {
		t.Fatalf("to play %s", g.ToPlay())
	}
	if err := g.Play(Point{4, 4}); err != nil {
		t.Fatal(err)
	}
	if g.At(Point{4, 4}) != Black {
		t.Fatal("stone not placed")
	}
	if g.ToPlay() != White {
		t.Fatalf("to play %s", g.ToPlay())
	}
}

func TestOccupiedAndOffBoard(t *testing.T) {
	g, _ := New(9, DefaultKomi)
	_ = g.Play(Point{0, 0})
	if err := g.Play(Point{0, 0}); !errors.Is(err, ErrOccupied) {
		t.Fatalf("got %v", err)
	}
	if err := g.Play(Point{-1, 0}); !errors.Is(err, ErrOffBoard) {
		t.Fatalf("got %v", err)
	}
	if err := g.Play(Point{9, 0}); !errors.Is(err, ErrOffBoard) {
		t.Fatalf("got %v", err)
	}
}

func TestSingleStoneCapture(t *testing.T) {
	g := load(t, 9, Black,
		".........",
		".........",
		"...X.....",
		"..XO.....",
		"...X.....",
		".........",
		".........",
		".........",
		".........",
	)
	// White at (3,3) has one liberty at (4,3). Black fills it.
	if err := g.Play(Point{4, 3}); err != nil {
		t.Fatal(err)
	}
	if g.At(Point{3, 3}) != Empty {
		t.Fatalf("expected capture, board:\n%s", grid(g))
	}
	if g.Captured(Black) != 1 {
		t.Fatalf("captured %d", g.Captured(Black))
	}
}

func TestMultiGroupCapture(t *testing.T) {
	// One black stone captures two separate white stones.
	g := load(t, 9, Black,
		".........",
		".........",
		"...X.X...",
		"..XO.OX..",
		"...X.X...",
		".........",
		".........",
		".........",
		".........",
	)
	if err := g.Play(Point{4, 3}); err != nil {
		t.Fatal(err)
	}
	if g.At(Point{3, 3}) != Empty || g.At(Point{5, 3}) != Empty {
		t.Fatalf("expected both captured:\n%s", grid(g))
	}
	if g.Captured(Black) != 2 {
		t.Fatalf("captured %d", g.Captured(Black))
	}
}

func TestSuicideRejected(t *testing.T) {
	g := load(t, 9, White,
		".........",
		".........",
		"...X.....",
		"..X.X....",
		"...X.....",
		".........",
		".........",
		".........",
		".........",
	)
	if err := g.Play(Point{3, 3}); !errors.Is(err, ErrSuicide) {
		t.Fatalf("got %v, want suicide", err)
	}
	if g.At(Point{3, 3}) != Empty {
		t.Fatal("suicide placed a stone")
	}
}

func TestCapturingSuicideAllowed(t *testing.T) {
	// White fills the last liberty of a surrounded black group; the capture
	// leaves the played stone with liberties.
	g := load(t, 9, White,
		".........",
		"..OOOOO..",
		"..OXXXO..",
		"..OX.XO..",
		"..OXXXO..",
		"..OOOOO..",
		".........",
		".........",
		".........",
	)
	if err := g.Play(Point{4, 3}); err != nil {
		t.Fatal(err)
	}
	if g.At(Point{3, 2}) != Empty {
		t.Fatalf("black group should be gone:\n%s", grid(g))
	}
	if g.At(Point{4, 3}) != White {
		t.Fatal("white stone should stay")
	}
	if g.Captured(White) != 8 {
		t.Fatalf("captured %d, want 8", g.Captured(White))
	}
}

func TestSnapback(t *testing.T) {
	// White captures a one-stone throw-in, then Black recaptures the whole
	// white group — more than one stone, so it is not ko.
	g := load(t, 9, White,
		".........",
		"..XXXXX..",
		"..XOOOX..",
		"..XOX.X..",
		"..XOOOX..",
		"..XXXXX..",
		".........",
		".........",
		".........",
	)
	if err := g.Play(Point{5, 3}); err != nil {
		t.Fatal(err)
	}
	if g.At(Point{4, 3}) != Empty {
		t.Fatalf("white should capture the throw-in:\n%s", grid(g))
	}
	if g.Ko() != nil {
		t.Fatalf("snapback setup should not set ko, got %v", g.Ko())
	}
	if err := g.Play(Point{4, 3}); err != nil {
		t.Fatal(err)
	}
	if g.At(Point{3, 2}) != Empty || g.At(Point{5, 3}) != Empty {
		t.Fatalf("black should snap back the white group:\n%s", grid(g))
	}
	if g.Captured(Black) < 2 {
		t.Fatalf("snapback should take more than one stone, got %d", g.Captured(Black))
	}
}

func TestSimpleKo(t *testing.T) {
	g := load(t, 9, Black,
		".........",
		".........",
		"...XO....",
		"..XO.O...",
		"...XO....",
		".........",
		".........",
		".........",
		".........",
	)
	if err := g.Play(Point{4, 3}); err != nil {
		t.Fatal(err)
	}
	if g.At(Point{3, 3}) != Empty {
		t.Fatalf("ko capture failed:\n%s", grid(g))
	}
	ko := g.Ko()
	if ko == nil || *ko != (Point{3, 3}) {
		t.Fatalf("ko point = %v, want (3,3)", ko)
	}
	if err := g.Play(Point{3, 3}); !errors.Is(err, ErrKo) {
		t.Fatalf("immediate recapture got %v, want ko", err)
	}
	if err := g.Play(Point{0, 0}); err != nil {
		t.Fatal(err)
	}
	if g.Ko() != nil {
		t.Fatal("ko should clear after a non-recapture")
	}
	if err := g.Play(Point{8, 8}); err != nil {
		t.Fatal(err)
	}
	if err := g.Play(Point{3, 3}); err != nil {
		t.Fatalf("recapture on the next turn: %v", err)
	}
	if g.At(Point{4, 3}) != Empty {
		t.Fatalf("white should take back the ko:\n%s", grid(g))
	}
}

func TestTwoPassesEndGame(t *testing.T) {
	g, _ := New(9, DefaultKomi)
	if err := g.Pass(); err != nil {
		t.Fatal(err)
	}
	if g.Over() {
		t.Fatal("one pass should not end the game")
	}
	if err := g.Pass(); err != nil {
		t.Fatal(err)
	}
	if !g.Over() {
		t.Fatal("two passes should end the game")
	}
	if err := g.Play(Point{0, 0}); !errors.Is(err, ErrGameOver) {
		t.Fatalf("got %v", err)
	}
}

func TestResign(t *testing.T) {
	g, _ := New(9, DefaultKomi)
	if err := g.Resign(); err != nil {
		t.Fatal(err)
	}
	if !g.Over() || g.Resigned() != Black {
		t.Fatalf("over=%v resigned=%s", g.Over(), g.Resigned())
	}
	r := g.Score(nil)
	if r.Winner != White || r.Resigned != Black {
		t.Fatalf("result %+v", r)
	}
}

func TestUndoRestoresState(t *testing.T) {
	g := load(t, 9, Black,
		".........",
		".........",
		"...X.....",
		"..XO.....",
		"...X.....",
		".........",
		".........",
		".........",
		".........",
	)
	before := grid(g)
	toPlay := g.ToPlay()
	if err := g.Play(Point{4, 3}); err != nil {
		t.Fatal(err)
	}
	if g.Captured(Black) != 1 {
		t.Fatal("expected a capture before undo")
	}
	if err := g.Undo(); err != nil {
		t.Fatal(err)
	}
	if grid(g) != before {
		t.Fatalf("board after undo:\n%s\nwant:\n%s", grid(g), before)
	}
	if g.ToPlay() != toPlay {
		t.Fatalf("to play %s", g.ToPlay())
	}
	if g.Captured(Black) != 0 {
		t.Fatalf("captures %d", g.Captured(Black))
	}
	if g.Ko() != nil {
		t.Fatal("ko should be cleared")
	}
}

func TestUndoPassAndEmptyHistory(t *testing.T) {
	g, _ := New(9, DefaultKomi)
	if err := g.Undo(); !errors.Is(err, ErrNothingUndo) {
		t.Fatalf("got %v", err)
	}
	_ = g.Pass()
	_ = g.Pass()
	if !g.Over() {
		t.Fatal("expected over")
	}
	if err := g.Undo(); err != nil {
		t.Fatal(err)
	}
	if g.Over() {
		t.Fatal("undo of second pass should reopen the game")
	}
}
