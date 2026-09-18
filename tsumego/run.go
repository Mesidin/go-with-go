package tsumego

import "go-with-go/engine"

// Run is an in-progress problem set.
type Run struct {
	problems []Problem
	index    int
	game     *engine.Game
	line     []engine.Point
	ply      int
	solved   bool
	hint     string
}

func NewRun() (*Run, error) {
	r := &Run{problems: Problems()}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Run) load() error {
	pr := r.problems[r.index]
	g, err := engine.Load(pr.Size, engine.DefaultKomi, pr.ToPlay, pr.Grid...)
	if err != nil {
		return err
	}
	r.game = g
	r.line = nil
	r.ply = 0
	r.solved = false
	r.hint = ""
	return nil
}

func (r *Run) Game() *engine.Game { return r.game }
func (r *Run) Problem() Problem   { return r.problems[r.index] }
func (r *Run) Index() int         { return r.index }
func (r *Run) Count() int         { return len(r.problems) }
func (r *Run) Solved() bool       { return r.solved }
func (r *Run) Hint() string       { return r.hint }
func (r *Run) Marks() []engine.Point {
	return r.Problem().Marks
}

func (r *Run) Prompt() string {
	if r.solved {
		return "Correct. " + r.Problem().Explain
	}
	if r.hint != "" {
		return r.hint
	}
	return r.Problem().Prompt
}

func matchingLine(lines [][]engine.Point, p engine.Point) []engine.Point {
	for _, ln := range lines {
		if len(ln) > 0 && ln[0] == p {
			return ln
		}
	}
	return nil
}

func (r *Run) Click(p engine.Point) (flash bool) {
	if r.solved || r.game == nil {
		return false
	}
	pr := r.Problem()
	if r.line == nil {
		ln := matchingLine(pr.Lines, p)
		if ln == nil {
			if err := r.game.Play(p); err != nil {
				r.hint = "Not the vital point. Try a marked idea."
				return true
			}
			_ = r.game.Undo()
			r.hint = "Legal, but not the solution. Try again."
			return true
		}
		if pr.AcceptIllegal {
			r.solved = true
			r.hint = ""
			return true
		}
		if err := r.game.Play(p); err != nil {
			r.hint = err.Error()
			return true
		}
		r.line = ln
		r.ply = 1
		r.autoReply()
		return false
	}
	if r.ply >= len(r.line) || r.line[r.ply] != p {
		r.hint = "Continue the sequence on the marked idea."
		return true
	}
	if err := r.game.Play(p); err != nil {
		r.hint = err.Error()
		return true
	}
	r.ply++
	r.autoReply()
	return false
}

func (r *Run) autoReply() {
	if r.ply >= len(r.line) {
		r.solved = true
		r.hint = ""
		return
	}
	// Opponent ply.
	reply := r.line[r.ply]
	if err := r.game.Play(reply); err != nil {
		r.solved = true
		return
	}
	r.ply++
	if r.ply >= len(r.line) {
		r.solved = true
		r.hint = ""
	}
}

func (r *Run) Next() error {
	r.index++
	if r.index >= len(r.problems) {
		r.index = 0
	}
	return r.load()
}

func (r *Run) Prev() error {
	if r.index > 0 {
		r.index--
	}
	return r.load()
}

func (r *Run) Reset() error { return r.load() }
