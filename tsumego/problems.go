package tsumego

import "go-with-go/engine"

// Problem is a life-and-death exercise. Lines are winning sequences
// starting with the student's move; even indices are the student.
type Problem struct {
	Title          string
	Prompt         string
	Explain        string
	Rank           string
	Size           int
	ToPlay         engine.Color
	Grid           []string
	Marks          []engine.Point
	Lines          [][]engine.Point
	AcceptIllegal  bool
}

// AcceptIllegal is true when the student is meant to try a suicide (eyes).
// Problems is a short beginner set. Rank is easy or club.

func Problems() []Problem {
	return []Problem{
		{
			Title:  "Take the stone",
			Prompt: "White is in atari. Capture it.",
			Explain: "The last liberty is the marked empty point. Filling it takes the stone.",
			Rank:   "easy",
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
			Marks: []engine.Point{{X: 3, Y: 3}, {X: 4, Y: 3}},
			Lines: [][]engine.Point{{{X: 4, Y: 3}}},
		},
		{
			Title:  "Capture the pair",
			Prompt: "Two white stones share one liberty. Take the group.",
			Explain: "Connected stones share liberties. One play captures both.",
			Rank:   "easy",
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
			Marks: []engine.Point{{X: 3, Y: 3}, {X: 4, Y: 3}, {X: 5, Y: 3}},
			Lines: [][]engine.Point{{{X: 5, Y: 3}}},
		},
		{
			Title:   "Snapback",
			Prompt:  "White can capture a throw-in, then you recapture the group.",
			Explain: "Let White take the one stone; the capturing group is then in atari. Recapture is not ko because you take more than one stone.",
			Rank:    "club",
			Size:    9,
			ToPlay:  engine.White,
			Grid: []string{
				".........",
				"..XXXXX..",
				"..XOOOX..",
				"..XOX.X..",
				"..XOOOX..",
				"..XXXXX..",
				".........",
				".........",
				".........",
			},
			Marks: []engine.Point{{X: 4, Y: 3}, {X: 5, Y: 3}},
			Lines: [][]engine.Point{
				{{X: 5, Y: 3}, {X: 4, Y: 3}},
			},
		},
		{
			Title:          "Two eyes live",
			Prompt:         "Black already lives with two eyes. Click either eye — the play is suicide.",
			Explain:        "A group with two real eyes cannot be captured. Illegal moves are not placed.",
			Rank:           "easy",
			Size:           9,
			ToPlay:         engine.White,
			AcceptIllegal:  true,
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
			Marks: []engine.Point{{X: 2, Y: 2}, {X: 4, Y: 2}},
			Lines: [][]engine.Point{
				{{X: 2, Y: 2}},
				{{X: 4, Y: 2}},
			},
		},
		{
			Title:   "Make two eyes",
			Prompt:  "Black to live. Play the center of the three-point space.",
			Explain: "The middle point splits the space into two eyes. A side point leaves a false eye.",
			Rank:    "club",
			Size:    9,
			ToPlay:  engine.Black,
			Grid: []string{
				".........",
				".OOOOO...",
				".OXXXO...",
				".O...O...",
				".OXXXO...",
				".OOOOO...",
				".........",
				".........",
				".........",
			},
			Marks: []engine.Point{{X: 2, Y: 3}, {X: 3, Y: 3}, {X: 4, Y: 3}},
			Lines: [][]engine.Point{{{X: 3, Y: 3}}},
		},
		{
			Title:   "Net",
			Prompt:  "The white stone cannot run. Close the net.",
			Explain: "A net (geta) captures by surrounding the escape routes, not by filling liberties yet.",
			Rank:    "club",
			Size:    9,
			ToPlay:  engine.Black,
			Grid: []string{
				".........",
				".........",
				"...X.....",
				"....O....",
				"..X......",
				".........",
				".........",
				".........",
				".........",
			},
			Marks: []engine.Point{{X: 4, Y: 3}},
			Lines: [][]engine.Point{{{X: 5, Y: 5}}, {{X: 6, Y: 4}}},
		},
		{
			Title:   "Corner kill",
			Prompt:  "White is cramped in the corner. Take the vital point.",
			Explain: "The 2-2 point (or the marked cut) is often the killing play in a small corner.",
			Rank:    "club",
			Size:    9,
			ToPlay:  engine.Black,
			Grid: []string{
				"OOX......",
				"O........",
				"X........",
				".........",
				".........",
				".........",
				".........",
				".........",
				".........",
			},
			Marks: []engine.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}},
			Lines: [][]engine.Point{{{X: 1, Y: 1}}},
		},
		{
			Title:   "Ko capture",
			Prompt:  "Take the ko. You cannot recapture immediately afterward.",
			Explain: "Capturing the single stone starts a ko. White may not take back on the next play.",
			Rank:    "club",
			Size:    9,
			ToPlay:  engine.Black,
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
			Marks: []engine.Point{{X: 3, Y: 3}, {X: 4, Y: 3}},
			Lines: [][]engine.Point{{{X: 4, Y: 3}}},
		},
	}
}
