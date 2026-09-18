package bot

import (
	"math/rand"
	"testing"

	"go-with-go/engine"
)

func TestPickOnlyLegal(t *testing.T) {
	g, err := engine.New(9, engine.DefaultKomi)
	if err != nil {
		t.Fatal(err)
	}
	h := New(rand.New(rand.NewSource(42)))
	for i := 0; i < 40; i++ {
		if g.Over() {
			break
		}
		p, ok := h.Pick(g)
		if !ok {
			if err := g.Pass(); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if !g.Legal(p) {
			t.Fatalf("bot chose illegal %v", p)
		}
		if err := g.Play(p); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNeverFillsOwnTrueEyeWhenOtherMovesExist(t *testing.T) {
	g, err := engine.New(9, engine.DefaultKomi)
	if err != nil {
		t.Fatal(err)
	}
	// Black 1-point eye at (1,1), lots of empty rest of board.
	plays := []engine.Point{
		{X: 1, Y: 0}, {X: 0, Y: 1}, {X: 2, Y: 1}, {X: 1, Y: 2},
	}
	for _, p := range plays {
		if err := g.Play(p); err != nil {
			t.Fatal(err)
		}
		// White dumps stones far away so the eye stays Black's.
		far := engine.Point{X: 8, Y: 8}
		for !g.Legal(far) {
			far.X--
		}
		if err := g.Play(far); err != nil {
			t.Fatal(err)
		}
	}
	if g.ToPlay() != engine.Black {
		t.Fatal("black to play")
	}
	h := New(rand.New(rand.NewSource(1)))
	for i := 0; i < 20; i++ {
		p, ok := h.Pick(g)
		if !ok {
			t.Fatal("bot passed with plenty of board left")
		}
		if p == (engine.Point{X: 1, Y: 1}) {
			t.Fatalf("filled own eye at %v", p)
		}
		_ = g.Play(p)
		if g.Over() {
			break
		}
		if q, ok := h.Pick(g); ok {
			_ = g.Play(q)
		} else {
			_ = g.Pass()
		}
	}
}

func TestPrefersCapture(t *testing.T) {
	g := captureSetup(t)
	h := New(rand.New(rand.NewSource(7)))
	p, ok := h.Pick(g)
	if !ok {
		t.Fatal("pass")
	}
	if p != (engine.Point{X: 4, Y: 3}) {
		t.Fatalf("expected capture at (4,3), got %v", p)
	}
}

func captureSetup(t *testing.T) *engine.Game {
	t.Helper()
	g, err := engine.New(9, engine.DefaultKomi)
	if err != nil {
		t.Fatal(err)
	}
	// Black surrounds (3,3) on three sides; White sits in the hole with one liberty at (4,3).
	moves := []engine.Point{
		{X: 3, Y: 2}, {X: 3, Y: 3},
		{X: 2, Y: 3}, {X: 8, Y: 8},
		{X: 3, Y: 4}, {X: 8, Y: 0},
	}
	for _, p := range moves {
		if err := g.Play(p); err != nil {
			t.Fatalf("play %v: %v", p, err)
		}
	}
	if g.ToPlay() != engine.Black {
		t.Fatal("black to play")
	}
	return g
}
