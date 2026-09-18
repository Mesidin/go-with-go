package bot

import (
	"math/rand"

	"go-with-go/engine"
)

// Heuristic is a capture-aware legal-move bot. It is a 9x9 sparring partner,
// not a ranked engine.
type Heuristic struct {
	rng *rand.Rand
}

func New(rng *rand.Rand) *Heuristic {
	if rng == nil {
		rng = rand.New(rand.NewSource(1))
	}
	return &Heuristic{rng: rng}
}

// Pick chooses a placement. ok is false when the bot should pass (no legal
// non-eye-fill move, or no legal move at all).
func (h *Heuristic) Pick(g *engine.Game) (engine.Point, bool) {
	moves := g.LegalMoves()
	if len(moves) == 0 {
		return engine.Point{}, false
	}
	me := g.ToPlay()
	var nonEye []engine.Point
	for _, p := range moves {
		if !isTrueEye(g, p, me) {
			nonEye = append(nonEye, p)
		}
	}
	if len(nonEye) == 0 {
		return engine.Point{}, false
	}

	bestScore := -1e18
	var best []engine.Point
	for _, p := range nonEye {
		s := h.scoreMove(g, p)
		if s > bestScore {
			bestScore = s
			best = []engine.Point{p}
		} else if s == bestScore {
			best = append(best, p)
		}
	}
	return best[h.rng.Intn(len(best))], true
}

func (h *Heuristic) scoreMove(g *engine.Game, p engine.Point) float64 {
	me := g.ToPlay()
	opp := me.Opponent()
	beforeCap := g.Captured(me)

	trial := g.Clone()
	if err := trial.Play(p); err != nil {
		return -1e9
	}
	gained := trial.Captured(me) - beforeCap

	s := float64(gained) * 100
	s += h.rng.Float64() * 3 // jitter so games vary

	if gained == 0 && isTrueEye(g, p, opp) {
		// Taking the opponent's eye is usually filling a false or real eye;
		// still slightly useful to steal a point, but don't prefer it over
		// real plays.
		s -= 20
	}

	// Saving own groups in atari: a neighbor of p was our stone with 1 liberty.
	for _, n := range g.Neighbors(p) {
		if g.At(n) == me && g.LibertyCount(n) == 1 {
			if trial.At(n) == me && trial.LibertyCount(n) > 1 {
				s += 80
			}
		}
		if g.At(n) == opp && g.LibertyCount(n) == 2 {
			// Approach a group that will be in atari after this (if not captured).
			if trial.At(n) == opp && trial.LibertyCount(n) == 1 {
				s += 40
			}
		}
	}

	// Opening bias: star points, then closer to center.
	size := g.Size()
	cx, cy := float64(size-1)/2, float64(size-1)/2
	dist := abs(float64(p.X)-cx) + abs(float64(p.Y)-cy)
	s += 8 - dist*0.4
	for _, hpt := range engine.Hoshi(size) {
		if hpt == p {
			s += 12
		}
	}
	return s
}

func isTrueEye(g *engine.Game, p engine.Point, c engine.Color) bool {
	if g.At(p) != engine.Empty {
		return false
	}
	n := 0
	for _, d := range [4]engine.Point{{X: 0, Y: -1}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: -1, Y: 0}} {
		q := engine.Point{X: p.X + d.X, Y: p.Y + d.Y}
		if q.X < 0 || q.Y < 0 || q.X >= g.Size() || q.Y >= g.Size() {
			n++
			continue
		}
		if g.At(q) != c {
			return false
		}
		n++
	}
	return n > 0
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
