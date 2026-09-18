package learn

import (
	"testing"

	"go-with-go/engine"
)

func TestLessonsLoadAndTargetsAreConsistent(t *testing.T) {
	for i, les := range Lessons() {
		g, err := engine.Load(les.Size, engine.DefaultKomi, les.ToPlay, les.Grid...)
		if err != nil {
			t.Fatalf("lesson %d %q: load: %v", i, les.Title, err)
		}
		if len(les.Steps) == 0 {
			t.Fatalf("lesson %d %q: no steps", i, les.Title)
		}
		for s, st := range les.Steps {
			switch st.Kind {
			case KindPlay:
				if st.AllowAny {
					if len(g.LegalMoves()) == 0 {
						t.Fatalf("lesson %d step %d: AllowAny but no legal moves", i, s)
					}
					continue
				}
				if len(st.Targets) == 0 {
					t.Fatalf("lesson %d step %d: play step has no targets", i, s)
				}
				for _, p := range st.Targets {
					if !g.Legal(p) {
						t.Fatalf("lesson %d %q step %d: target %v is not legal", i, les.Title, s, p)
					}
				}
			case KindTryIllegal:
				if len(st.Try) == 0 {
					t.Fatalf("lesson %d step %d: try-illegal has no points", i, s)
				}
				for _, p := range st.Try {
					if g.Legal(p) {
						t.Fatalf("lesson %d %q step %d: %v should be illegal", i, les.Title, s, p)
					}
				}
			case KindPass:
				// checked by the walkthrough; loading the board is enough here
			}
		}
	}
}

func TestRunCaptureLesson(t *testing.T) {
	r, err := NewRun()
	if err != nil {
		t.Fatal(err)
	}
	// Skip to Capture (index 3).
	r.index = 3
	if err := r.load(); err != nil {
		t.Fatal(err)
	}
	r.Click(engine.Point{X: 0, Y: 0})
	if r.Complete() {
		t.Fatal("wrong point should not finish the lesson")
	}
	r.Click(engine.Point{X: 4, Y: 3})
	if !r.Complete() {
		t.Fatal("capture should finish the lesson")
	}
	if r.Game().At(engine.Point{X: 3, Y: 3}) != engine.Empty {
		t.Fatal("white stone should be captured")
	}
}

func TestRunWalkthrough(t *testing.T) {
	r, err := NewRun()
	if err != nil {
		t.Fatal(err)
	}
	guard := 0
	for !r.Graduated() {
		guard++
		if guard > 80 {
			t.Fatal("walkthrough stuck")
		}
		if r.Complete() {
			if err := r.NextLesson(); err != nil {
				t.Fatal(err)
			}
			continue
		}
		st := r.Step()
		switch st.Kind {
		case KindPlay:
			var p engine.Point
			if st.AllowAny {
				moves := r.Game().LegalMoves()
				if len(moves) == 0 {
					t.Fatalf("no legal moves in %q", r.Lesson().Title)
				}
				p = moves[0]
			} else {
				p = st.Targets[0]
			}
			r.Click(p)
		case KindTryIllegal:
			r.Click(st.Try[0])
		case KindPass:
			if !r.Pass() {
				t.Fatalf("pass failed in %q: %s", r.Lesson().Title, r.Hint())
			}
		}
	}
}

func TestRunSuicideThenAdvance(t *testing.T) {
	r, err := NewRun()
	if err != nil {
		t.Fatal(err)
	}
	r.index = 5
	if err := r.load(); err != nil {
		t.Fatal(err)
	}
	flash := r.Click(engine.Point{X: 3, Y: 3})
	if !flash {
		t.Fatal("suicide click should flash")
	}
	if !r.Complete() {
		t.Fatal("trying the illegal eye/hole should complete the step")
	}
	if r.Game().At(engine.Point{X: 3, Y: 3}) != engine.Empty {
		t.Fatal("suicide must not place a stone")
	}
}
