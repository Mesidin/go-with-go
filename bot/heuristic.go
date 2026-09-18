package bot

import (
	"math/rand"

	"go-with-go/engine"
)

// Heuristic is a capture-aware legal-move bot.
// Easy is a 9×9 sparring partner. Club adds self-atari and 1-reply capture checks.
type Heuristic struct {
	rng    *rand.Rand
	Strong bool
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

	ownN := 0
	for _, n := range g.Neighbors(p) {
		if g.At(n) == me {
			ownN++
		}
	}
	s += float64(ownN) * 6

	if h.Strong {
		if gained == 0 && trial.LibertyCount(p) == 1 {
			s -= 70 // self-atari
		}
		if n := maxReplyCapture(trial, opp); n > 0 {
			s -= float64(n) * 90
		}
	}
	return s
}

func maxReplyCapture(g *engine.Game, opp engine.Color) int {
	// Cheap: if any of our groups has one liberty, opponent can take it
	// by playing that liberty (when legal).
	me := opp.Opponent()
	seen := map[engine.Point]bool{}
	best := 0
	size := g.Size()
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			p := engine.Point{X: x, Y: y}
			if g.At(p) != me || seen[p] {
				continue
			}
			grp := g.Group(p)
			for _, q := range grp {
				seen[q] = true
			}
			if g.LibertyCount(p) != 1 {
				continue
			}
			libs := libertiesOf(g, p)
			if len(libs) != 1 {
				continue
			}
			if !g.Legal(libs[0]) {
				continue
			}
			if n := len(grp); n > best {
				best = n
			}
		}
	}
	return best
}

func libertiesOf(g *engine.Game, p engine.Point) []engine.Point {
	seen := map[engine.Point]struct{}{}
	var out []engine.Point
	for _, q := range g.Group(p) {
		for _, n := range g.Neighbors(q) {
			if g.At(n) != engine.Empty {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			out = append(out, n)
		}
	}
	return out
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
