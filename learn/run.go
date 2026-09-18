package learn

import (
	"go-with-go/engine"
)

// Run is an in-progress course through Lessons().
type Run struct {
	lessons   []Lesson
	index     int
	step      int
	game      *engine.Game
	hint      string
	complete  bool // current lesson's steps are done
	graduated bool // every lesson finished
}

func NewRun() (*Run, error) {
	r := &Run{lessons: Lessons()}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Run) load() error {
	if r.index >= len(r.lessons) {
		r.graduated = true
		r.complete = true
		r.game = nil
		return nil
	}
	les := r.lessons[r.index]
	g, err := engine.Load(les.Size, engine.DefaultKomi, les.ToPlay, les.Grid...)
	if err != nil {
		return err
	}
	r.game = g
	r.step = 0
	r.hint = ""
	r.complete = false
	return nil
}

func (r *Run) Game() *engine.Game { return r.game }
func (r *Run) Hint() string       { return r.hint }
func (r *Run) Complete() bool     { return r.complete }
func (r *Run) Graduated() bool    { return r.graduated }
func (r *Run) Index() int         { return r.index }
func (r *Run) Count() int         { return len(r.lessons) }

func (r *Run) Lesson() Lesson {
	if r.index >= len(r.lessons) {
		return Lesson{Title: "You know enough to play", Done: "Start a 9×9 game from the menu."}
	}
	return r.lessons[r.index]
}

func (r *Run) Step() Step {
	les := r.Lesson()
	if r.complete || r.step >= len(les.Steps) {
		return Step{}
	}
	return les.Steps[r.step]
}

func (r *Run) Prompt() string {
	if r.graduated {
		return r.Lesson().Done
	}
	if r.complete {
		return r.Lesson().Done
	}
	if h := r.hint; h != "" {
		return h
	}
	return r.Step().Prompt
}

func (r *Run) Marks() []engine.Point {
	if r.complete || r.graduated {
		return nil
	}
	return r.Step().Marks()
}

func contains(pts []engine.Point, p engine.Point) bool {
	for _, q := range pts {
		if q == p {
			return true
		}
	}
	return false
}

func (r *Run) advance() {
	r.hint = ""
	r.step++
	if r.step >= len(r.Lesson().Steps) {
		r.complete = true
	}
}

// Click handles a board click. flash is true when the student should see a refusal.
func (r *Run) Click(p engine.Point) (flash bool) {
	if r.game == nil || r.complete || r.graduated {
		return false
	}
	st := r.Step()
	switch st.Kind {
	case KindTryIllegal:
		if contains(st.Try, p) {
			_ = r.game.Play(p) // expected to fail
			r.advance()
			return true
		}
		r.hint = st.Hint
		if r.hint == "" {
			r.hint = "Click a marked point."
		}
		return false
	case KindPass:
		r.hint = st.Hint
		if r.hint == "" {
			r.hint = "Pass to continue (P)."
		}
		return false
	case KindPlay:
		if !st.AllowAny && !contains(st.Targets, p) {
			r.hint = st.Hint
			if r.hint == "" {
				r.hint = "Play on a marked point."
			}
			return true
		}
		if err := r.game.Play(p); err != nil {
			r.hint = err.Error()
			return true
		}
		r.advance()
		return false
	}
	return false
}

// Pass handles a pass. ok is false if this step does not want a pass.
func (r *Run) Pass() (ok bool) {
	if r.game == nil || r.complete || r.graduated {
		return false
	}
	st := r.Step()
	if st.Kind != KindPass {
		r.hint = "Don't pass yet — follow the prompt."
		return false
	}
	if err := r.game.Pass(); err != nil {
		r.hint = err.Error()
		return false
	}
	r.advance()
	return true
}

func (r *Run) NextLesson() error {
	if r.index < len(r.lessons)-1 {
		r.index++
		return r.load()
	}
	r.index = len(r.lessons)
	r.graduated = true
	r.complete = true
	r.game = r.game // keep last board visible
	return nil
}

func (r *Run) PrevLesson() error {
	if r.index <= 0 {
		return r.load()
	}
	r.index--
	r.graduated = false
	return r.load()
}
