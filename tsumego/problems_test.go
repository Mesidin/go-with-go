package tsumego

import (
	"testing"

	"go-with-go/engine"
)

func TestProblemsFirstMoves(t *testing.T) {
	for i, pr := range Problems() {
		g, err := engine.Load(pr.Size, engine.DefaultKomi, pr.ToPlay, pr.Grid...)
		if err != nil {
			t.Fatalf("%d %s: %v", i, pr.Title, err)
		}
		if len(pr.Lines) == 0 {
			t.Fatalf("%s: no lines", pr.Title)
		}
		for _, ln := range pr.Lines {
			if len(ln) == 0 {
				t.Fatalf("%s: empty line", pr.Title)
			}
			ok := g.Legal(ln[0])
			if pr.AcceptIllegal && ok {
				t.Fatalf("%s: expected illegal first move %v", pr.Title, ln[0])
			}
			if !pr.AcceptIllegal && !ok {
				t.Fatalf("%s: first move %v not legal", pr.Title, ln[0])
			}
		}
	}
}

func TestRunSolvesCapture(t *testing.T) {
	r, err := NewRun()
	if err != nil {
		t.Fatal(err)
	}
	r.Click(engine.Point{X: 4, Y: 3})
	if !r.Solved() {
		t.Fatalf("hint %s", r.Hint())
	}
}
