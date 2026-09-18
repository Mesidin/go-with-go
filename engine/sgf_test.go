package engine

import (
	"strings"
	"testing"
)

func TestSGFRoundTrip(t *testing.T) {
	g, err := New(9, DefaultKomi)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Play(Point{X: 4, Y: 4}); err != nil {
		t.Fatal(err)
	}
	if err := g.Play(Point{X: 2, Y: 2}); err != nil {
		t.Fatal(err)
	}
	if err := g.Pass(); err != nil {
		t.Fatal(err)
	}
	raw := g.SGF()
	if !strings.Contains(raw, "SZ[9]") || !strings.Contains(raw, "KM[6.5]") {
		t.Fatalf("header: %s", raw)
	}
	got, err := ParseSGF(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Size() != 9 || got.MoveCount() != 3 {
		t.Fatalf("size %d moves %d", got.Size(), got.MoveCount())
	}
	if got.At(Point{X: 4, Y: 4}) != Black || got.At(Point{X: 2, Y: 2}) != White {
		t.Fatalf("stones not restored\n%s", raw)
	}
}

func TestSGFHandicap(t *testing.T) {
	g, err := New(9, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.PlaceHandicap(4); err != nil {
		t.Fatal(err)
	}
	if g.ToPlay() != White {
		t.Fatal("white should move first")
	}
	if g.At(Point{X: 2, Y: 6}) != Black {
		t.Fatal("expected a handicap stone")
	}
	got, err := ParseSGF(g.SGF())
	if err != nil {
		t.Fatal(err)
	}
	if got.Handicap() != 4 || got.ToPlay() != White {
		t.Fatalf("ha %d toPlay %s", got.Handicap(), got.ToPlay())
	}
}

func TestFormatPointSkipsI(t *testing.T) {
	if FormatPoint(Point{X: 8, Y: 0}, 9) != "J9" {
		t.Fatalf("got %s want J9", FormatPoint(Point{X: 8, Y: 0}, 9))
	}
	if FormatPoint(Point{X: 0, Y: 8}, 9) != "A1" {
		t.Fatalf("got %s want A1", FormatPoint(Point{X: 0, Y: 8}, 9))
	}
}
