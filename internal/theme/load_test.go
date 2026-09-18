package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGoldRushTOML(t *testing.T) {
	dir := t.TempDir()
	data := `accent = "#C9A227"
foreground = "#D9D9D9"
background = "#121212"
color8 = "#805B10"
`
	if err := os.WriteFile(filepath.Join(dir, "colors.toml"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	p, ok := loadFromDir(dir)
	if !ok {
		t.Fatal("parse failed")
	}
	if p.Hex(p.Accent) != "#c9a227" {
		t.Fatalf("accent %s", p.Hex(p.Accent))
	}
	if p.Hex(p.Background) != "#121212" {
		t.Fatalf("bg %s", p.Hex(p.Background))
	}
}

func TestDefaultIsGoldRush(t *testing.T) {
	t.Setenv("OMARCHY_THEME_DIR", filepath.Join(t.TempDir(), "missing"))
	p := resolve(GoldRush())
	if p.Name != "gold-rush" {
		t.Fatalf("name %s", p.Name)
	}
}
