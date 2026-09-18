package ui

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"go-with-go/engine"
)

func gamesDir() string {
	return "games"
}

func (a *App) saveSGF() {
	if a.eng == nil {
		return
	}
	if err := os.MkdirAll(gamesDir(), 0o755); err != nil {
		a.saveMsg = err.Error()
		return
	}
	name := time.Now().Format("20060102-150405") + ".sgf"
	path := filepath.Join(gamesDir(), name)
	if err := os.WriteFile(path, []byte(a.eng.SGF()+"\n"), 0o644); err != nil {
		a.saveMsg = err.Error()
		return
	}
	a.saveMsg = "Saved " + filepath.ToSlash(path)
}

func (a *App) openLoad() {
	a.refreshLoads()
	a.scene = sceneLoad
}

func (a *App) refreshLoads() {
	a.loadErr = ""
	a.loadFiles = nil
	ents, err := os.ReadDir(gamesDir())
	if err != nil {
		a.loadErr = "No games/ folder yet. Save a game first, or drop an .sgf file on the window."
		return
	}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".sgf") {
			a.loadFiles = append(a.loadFiles, e.Name())
		}
	}
	if len(a.loadFiles) == 0 {
		a.loadErr = "No .sgf files in games/. Save a game, or drop a file on the window."
	}
}

func (a *App) loadFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		a.loadErr = err.Error()
		return
	}
	g, err := engine.ParseSGF(string(data))
	if err != nil {
		a.loadErr = err.Error()
		return
	}
	a.eng = g
	a.size = g.Size()
	a.komi = g.Komi()
	a.handicap = g.Handicap()
	a.dead = map[engine.Point]struct{}{}
	a.saveMsg = "Loaded " + filepath.Base(path)
	a.resetClock()
	if g.Over() {
		a.finish()
		return
	}
	a.scene = scenePlay
}

func (a *App) handleDrops() {
	fsys := ebiten.DroppedFiles()
	if fsys == nil {
		return
	}
	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(d.Name()), ".sgf") {
			return nil
		}
		f, err := fsys.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			return nil
		}
		g, err := engine.ParseSGF(string(data))
		if err != nil {
			a.loadErr = err.Error()
			return fs.SkipAll
		}
		a.eng = g
		a.size = g.Size()
		a.komi = g.Komi()
		a.handicap = g.Handicap()
		a.dead = map[engine.Point]struct{}{}
		a.saveMsg = "Loaded " + d.Name()
		a.resetClock()
		if g.Over() {
			a.finish()
		} else {
			a.scene = scenePlay
		}
		return fs.SkipAll
	})
}

func (a *App) updateLoad(clicked bool, cx, cy int) {
	if !clicked {
		return
	}
	for _, b := range a.loadNav {
		if !b.contains(cx, cy) {
			continue
		}
		switch b.id {
		case "refresh":
			a.refreshLoads()
		case "menu":
			a.scene = sceneMenu
		}
		return
	}
	y := 160
	for _, name := range a.loadFiles {
		if cy >= y && cy < y+36 && cx > 80 && cx < screenW-80 {
			a.loadFile(filepath.Join(gamesDir(), name))
			return
		}
		y += 40
	}
}

func (a *App) drawLoad(screen *ebiten.Image, mx, my int) {
	drawLabel(screen, "Load SGF", screenW/2, 70, faceBig, colInk, true)
	drawLabel(screen, "Files in games/  ·  you can also drop an .sgf on the window", screenW/2, 114, faceSm, colMuted, true)
	if a.loadErr != "" {
		drawWrapped(screen, a.loadErr, 80, 160, screenW-160, 22, face, colWin)
	}
	y := 160
	for _, name := range a.loadFiles {
		b := button{80, float32(y), screenW - 160, 34, name, name}
		drawButton(screen, b, mx, my)
		y += 40
	}
	for _, b := range a.loadNav {
		drawButton(screen, b, mx, my)
	}
}

func (a *App) resetClock() {
	a.clockAt = time.Time{}
	if a.clockMin <= 0 {
		a.clockLeft = [3]time.Duration{}
		return
	}
	d := time.Duration(a.clockMin) * time.Minute
	a.clockLeft[engine.Black] = d
	a.clockLeft[engine.White] = d
}

func (a *App) tickClock() {
	if a.clockMin <= 0 || a.eng == nil || a.eng.Over() {
		a.clockAt = time.Time{}
		return
	}
	now := time.Now()
	if a.clockAt.IsZero() {
		a.clockAt = now
		return
	}
	dt := now.Sub(a.clockAt)
	a.clockAt = now
	c := a.eng.ToPlay()
	a.clockLeft[c] -= dt
	if a.clockLeft[c] <= 0 {
		a.clockLeft[c] = 0
		_ = a.eng.Resign()
		a.saveMsg = c.String() + " lost on time"
		a.finish()
	}
}

func (a *App) clockText() string {
	if a.clockMin <= 0 {
		return "no clock"
	}
	fmtDur := func(d time.Duration) string {
		if d < 0 {
			d = 0
		}
		s := int(d.Seconds())
		return fmt.Sprintf("%d:%02d", s/60, s%60)
	}
	return "B " + fmtDur(a.clockLeft[engine.Black]) + "   W " + fmtDur(a.clockLeft[engine.White])
}

func (a *App) drawMoveList(screen *ebiten.Image) {
	const x, w float32 = 680, 260
	y := float32(40)
	h := float32(screenH - hudH - 60)
	drawPanel(screen, x, y, w, h)
	vectorStrokeRect(screen, x, y, w, h, colWin)
	drawLabel(screen, "Moves", x+16, y+12, face, colWin, false)
	if a.eng == nil {
		return
	}
	row := y + 44
	for i, m := range a.eng.Moves() {
		if row > y+h-28 {
			drawLabel(screen, "…", x+16, row, faceSm, colMuted, false)
			break
		}
		line := fmt.Sprintf("%d. %s", i+1, m.Label(a.eng.Size()))
		drawLabel(screen, line, x+16, row, faceSm, colInk, false)
		row += 20
	}
}
