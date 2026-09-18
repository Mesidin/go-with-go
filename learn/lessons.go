package learn

import "go-with-go/engine"

// Kind is what the student must do to finish a step.
type Kind int

const (
	KindPlay Kind = iota
	KindPass
	KindTryIllegal
)

// Step is one prompt inside a lesson.
type Step struct {
	Prompt     string
	Kind       Kind
	Targets    []engine.Point // KindPlay: must play one of these, unless AllowAny
	AllowAny   bool
	Try        []engine.Point // KindTryIllegal: click one of these (move is refused)
	Hint       string
	Highlight  []engine.Point
}

func (s Step) Marks() []engine.Point {
	out := append([]engine.Point(nil), s.Highlight...)
	out = append(out, s.Targets...)
	out = append(out, s.Try...)
	return out
}

// Lesson is a short interactive explanation of one rule.
type Lesson struct {
	Title string
	Blurb string
	Size  int
	ToPlay engine.Color
	Grid  []string
	Steps []Step
	Done  string
}

// Lessons is the built-in beginner course, Japanese club rules.
func Lessons() []Lesson {
	empty := []string{
		".........",
		".........",
		".........",
		".........",
		".........",
		".........",
		".........",
		".........",
		".........",
	}
	return []Lesson{
		{
			Title:  "The board",
			Blurb:  "Go is played on the intersections, not in the squares. Black plays first.",
			Size:   9,
			ToPlay: engine.Black,
			Grid:   empty,
			Steps: []Step{{
				Prompt:   "Click any intersection to place a black stone.",
				Kind:     KindPlay,
				AllowAny: true,
			}},
			Done: "A game is a conversation: Black, then White, then Black again.",
		},
		{
			Title:  "Taking turns",
			Blurb:  "Players alternate. You cannot play on a point that already has a stone.",
			Size:   9,
			ToPlay: engine.White,
			Grid: []string{
				".........",
				".........",
				".........",
				".........",
				"....X....",
				".........",
				".........",
				".........",
				".........",
			},
			Steps: []Step{{
				Prompt:   "White's turn. Place a white stone on any empty point.",
				Kind:     KindPlay,
				AllowAny: true,
			}},
			Done: "Each stone stays where it is unless it is captured.",
		},
		{
			Title:  "Liberties",
			Blurb:  "A liberty is an empty point next to a stone: up, down, left, or right — not diagonal.",
			Size:   9,
			ToPlay: engine.White,
			Grid: []string{
				".........",
				".........",
				".........",
				".........",
				"....X....",
				".........",
				".........",
				".........",
				".........",
			},
			Steps: []Step{{
				Prompt: "This black stone has four liberties (marked). Play on one of them.",
				Kind:   KindPlay,
				Targets: []engine.Point{
					{X: 4, Y: 3}, {X: 5, Y: 4}, {X: 4, Y: 5}, {X: 3, Y: 4},
				},
				Highlight: []engine.Point{
					{X: 4, Y: 3}, {X: 5, Y: 4}, {X: 4, Y: 5}, {X: 3, Y: 4},
				},
				Hint: "Play on a marked liberty next to the black stone.",
			}},
			Done: "Fill all of a group's liberties and that group is captured.",
		},
		{
			Title:  "Capture",
			Blurb:  "A stone with one liberty is in atari. Playing on the last liberty captures it.",
			Size:   9,
			ToPlay: engine.Black,
			Grid: []string{
				".........",
				".........",
				"...X.....",
				"..XO.....",
				"...X.....",
				".........",
				".........",
				".........",
				".........",
			},
			Steps: []Step{{
				Prompt:    "Capture the white stone by playing on its last liberty.",
				Kind:      KindPlay,
				Targets:   []engine.Point{{X: 4, Y: 3}},
				Highlight: []engine.Point{{X: 4, Y: 3}},
				Hint:      "The marked empty point is White's last liberty.",
			}},
			Done: "Captured stones come off the board and become prisoners.",
		},
		{
			Title:  "Groups",
			Blurb:  "Connected stones of the same color share liberties. Capture the whole chain at once.",
			Size:   9,
			ToPlay: engine.Black,
			Grid: []string{
				".........",
				".........",
				"...XX....",
				"..XOO....",
				"...XX....",
				".........",
				".........",
				".........",
				".........",
			},
			Steps: []Step{{
				Prompt:    "These two white stones have one shared liberty. Capture the group.",
				Kind:      KindPlay,
				Targets:   []engine.Point{{X: 5, Y: 3}},
				Highlight: []engine.Point{{X: 5, Y: 3}},
				Hint:      "Play on the marked point to take both stones.",
			}},
			Done: "Always count liberties of the whole group, not each stone alone.",
		},
		{
			Title:  "Suicide",
			Blurb:  "You may not play a stone that would have no liberties — unless that play captures first.",
			Size:   9,
			ToPlay: engine.White,
			Grid: []string{
				".........",
				".........",
				"...X.....",
				"..X.X....",
				"...X.....",
				".........",
				".........",
				".........",
				".........",
			},
			Steps: []Step{{
				Prompt:    "Click the empty point in the middle. The move is suicide, so the game refuses it.",
				Kind:      KindTryIllegal,
				Try:       []engine.Point{{X: 3, Y: 3}},
				Highlight: []engine.Point{{X: 3, Y: 3}},
				Hint:      "Click the marked hole — White would have no liberties there.",
			}},
			Done: "The board stays empty: illegal moves are not placed.",
		},
		{
			Title:  "Capturing is not suicide",
			Blurb:  "If your play captures the opponent first, your stone gets liberties from the empty points left behind.",
			Size:   9,
			ToPlay: engine.White,
			Grid: []string{
				".........",
				"..OOOOO..",
				"..OXXXO..",
				"..OX.XO..",
				"..OXXXO..",
				"..OOOOO..",
				".........",
				".........",
				".........",
			},
			Steps: []Step{{
				Prompt:    "Play the marked point. White captures the black group and lives.",
				Kind:      KindPlay,
				Targets:   []engine.Point{{X: 4, Y: 3}},
				Highlight: []engine.Point{{X: 4, Y: 3}},
				Hint:      "Play in the middle of the black group.",
			}},
			Done: "Capture is checked before suicide. That is why this play is legal.",
		},
		{
			Title:  "Ko",
			Blurb:  "You cannot immediately recapture a single stone if that would repeat the position. That fight is called ko.",
			Size:   9,
			ToPlay: engine.Black,
			Grid: []string{
				".........",
				".........",
				"...XO....",
				"..XO.O...",
				"...XO....",
				".........",
				".........",
				".........",
				".........",
			},
			Steps: []Step{
				{
					Prompt:    "Capture the white stone by taking the ko.",
					Kind:      KindPlay,
					Targets:   []engine.Point{{X: 4, Y: 3}},
					Highlight: []engine.Point{{X: 4, Y: 3}},
					Hint:      "Play the marked point to capture.",
				},
				{
					Prompt:    "Now try to take it back immediately. The game will refuse — that is ko.",
					Kind:      KindTryIllegal,
					Try:       []engine.Point{{X: 3, Y: 3}},
					Highlight: []engine.Point{{X: 3, Y: 3}},
					Hint:      "Click the empty point you just captured.",
				},
				{
					Prompt:   "Play anywhere else. After a move elsewhere, the ko may be taken back.",
					Kind:     KindPlay,
					AllowAny: true,
					Hint:     "Play on any empty point that is not the ko.",
				},
			},
			Done: "To retake a ko, first play a threat elsewhere so the opponent answers.",
		},
		{
			Title:  "Two eyes",
			Blurb:  "A group with two separate eyes cannot be captured. Playing in a true eye is suicide.",
			Size:   9,
			ToPlay: engine.White,
			Grid: []string{
				".........",
				".XXXXX...",
				".X.X.X...",
				".XXXXX...",
				".........",
				".........",
				".........",
				".........",
				".........",
			},
			Steps: []Step{{
				Prompt:    "Click either eye inside the black group. Both plays are suicide.",
				Kind:      KindTryIllegal,
				Try:       []engine.Point{{X: 2, Y: 2}, {X: 4, Y: 2}},
				Highlight: []engine.Point{{X: 2, Y: 2}, {X: 4, Y: 2}},
				Hint:      "Click one of the two marked eyes.",
			}},
			Done: "One eye can be filled under atari; two eyes live. That is the basic life-and-death test.",
		},
		{
			Title:  "Territory and scoring",
			Blurb:  "Japanese scoring: empty points surrounded by one color, plus prisoners, plus 6.5 komi for White. Stones on the board do not count.",
			Size:   9,
			ToPlay: engine.Black,
			Grid: []string{
				"XXXXXXXXX",
				"X.......X",
				"X.......X",
				"X.......X",
				"XXXXXXXXX",
				"OOOOOOOOO",
				"O.......O",
				"O.......O",
				"OOOOOOOOO",
			},
			Steps: []Step{
				{
					Prompt: "Black has 21 territory, White 14. With 6.5 komi, Black wins by 0.5. Pass (button or P).",
					Kind:   KindPass,
					Hint:   "Press P or click Pass. Two consecutive passes end the game.",
				},
				{
					Prompt: "White passes too. Pass again to finish.",
					Kind:   KindPass,
					Hint:   "Pass once more.",
				},
			},
			Done: "Dame (empty points touching both colors) score for nobody. Mark dead groups before you confirm a real game.",
		},
	}
}
