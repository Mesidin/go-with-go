package engine

import "testing"

func TestJapaneseTerritoryAndKomi(t *testing.T) {
	g := load(t, 9, Black,
		"XXXXXXXXX",
		"X.......X",
		"X.......X",
		"X.......X",
		"XXXXXXXXX",
		"OOOOOOOOO",
		"O.......O",
		"O.......O",
		"OOOOOOOOO",
	)
	r := g.Score(nil)
	if r.BlackTerritory != 21 {
		t.Fatalf("black territory %d, want 21", r.BlackTerritory)
	}
	if r.WhiteTerritory != 14 {
		t.Fatalf("white territory %d, want 14", r.WhiteTerritory)
	}
	if r.BlackCaptures != 0 || r.WhiteCaptures != 0 {
		t.Fatalf("captures B=%d W=%d", r.BlackCaptures, r.WhiteCaptures)
	}
	if r.Komi != DefaultKomi {
		t.Fatalf("komi %v", r.Komi)
	}
	if r.Black != 21 {
		t.Fatalf("black score %v, want 21 (stones on the board must not count)", r.Black)
	}
	if r.White != 14+DefaultKomi {
		t.Fatalf("white score %v, want %v", r.White, 14+DefaultKomi)
	}
	if r.Winner != Black {
		t.Fatalf("winner %s", r.Winner)
	}
	if r.Margin() != r.Black-r.White {
		t.Fatalf("margin %v", r.Margin())
	}
	want := "Black wins by 0.5"
	if r.String() != want {
		t.Fatalf("headline %q, want %q", r.String(), want)
	}
}

func TestDameScoresForNeither(t *testing.T) {
	g := load(t, 9, Black,
		"XXXX.OOOO",
		"XXXX.OOOO",
		"XXXX.OOOO",
		"XXXX.OOOO",
		"XXXX.OOOO",
		"XXXX.OOOO",
		"XXXX.OOOO",
		"XXXX.OOOO",
		"XXXX.OOOO",
	)
	r := g.Score(nil)
	if r.BlackTerritory != 0 || r.WhiteTerritory != 0 {
		t.Fatalf("dame should not be territory: B=%d W=%d", r.BlackTerritory, r.WhiteTerritory)
	}
	if r.Winner != White {
		t.Fatalf("with empty territory, 6.5 komi should give White the win, got %s", r.Winner)
	}
}

func TestDeadStonesBecomePrisonersAndTerritory(t *testing.T) {
	g := load(t, 9, Black,
		"XXXXXXXXX",
		"X.......X",
		"X...O...X",
		"X.......X",
		"XXXXXXXXX",
		"OOOOOOOOO",
		"O.......O",
		"O.......O",
		"OOOOOOOOO",
	)
	alive := g.Score(nil)
	if alive.BlackTerritory != 0 {
		t.Fatalf("an unmarked white stone inside black's box should kill the territory, got %d", alive.BlackTerritory)
	}

	dead := map[Point]struct{}{{4, 2}: {}}
	r := g.Score(dead)
	if r.BlackCaptures != 1 {
		t.Fatalf("dead white stone should be a black prisoner, got %d", r.BlackCaptures)
	}
	if r.BlackTerritory != 21 {
		t.Fatalf("black territory after marking dead %d, want 21", r.BlackTerritory)
	}
	if r.Black != 22 {
		t.Fatalf("black score %v, want 22 (21 territory + 1 prisoner)", r.Black)
	}
}

func TestCapturesAddedToJapaneseScore(t *testing.T) {
	g, _ := New(9, DefaultKomi)
	g.captured[Black] = 3
	g.captured[White] = 1
	// Empty board: the whole board is dame-less empty touching nobody, so
	// no territory (region touches neither color).
	r := g.Score(nil)
	if r.Black != 3 {
		t.Fatalf("black %v, want 3 prisoners", r.Black)
	}
	if r.White != 1+DefaultKomi {
		t.Fatalf("white %v", r.White)
	}
}

func TestDefaultKomiIsJapanese(t *testing.T) {
	if DefaultKomi != 6.5 {
		t.Fatalf("DefaultKomi = %v, want 6.5", DefaultKomi)
	}
	for _, size := range ValidSizes {
		g, err := New(size, DefaultKomi)
		if err != nil {
			t.Fatal(err)
		}
		if g.Komi() != 6.5 {
			t.Fatalf("size %d komi %v", size, g.Komi())
		}
	}
}

func TestResumeAfterPasses(t *testing.T) {
	g, _ := New(9, DefaultKomi)
	_ = g.Pass()
	_ = g.Pass()
	if err := g.ResumeAfterPasses(); err != nil {
		t.Fatal(err)
	}
	if g.Over() {
		t.Fatal("should be resumable")
	}
	if err := g.Play(Point{0, 0}); err != nil {
		t.Fatal(err)
	}
}
