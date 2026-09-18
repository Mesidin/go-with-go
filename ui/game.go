package ui

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"go-with-go/bot"
	"go-with-go/engine"
	"go-with-go/internal/meta"
	"go-with-go/internal/theme"
	"go-with-go/learn"
	"go-with-go/tsumego"
)

const (
	screenW = 960
	screenH = 1040
	hudH = 168
)

type scene int

const (
	sceneMenu scene = iota
	sceneSetup
	scenePlay
	sceneScore
	sceneResult
	sceneLearn
	sceneTsumego
	sceneLoad
)

// Mode is who plays each color.
type Mode int

const (
	HumanHuman Mode = iota
	HumanBlack
	HumanWhite
)

func (m Mode) String() string {
	switch m {
	case HumanBlack:
		return "Black vs bot"
	case HumanWhite:
		return "White vs bot"
	default:
		return "Humans"
	}
}

func (m Mode) vsBot() bool { return m == HumanBlack || m == HumanWhite }

// App is the Ebitengine front end.
type App struct {
	scene  scene
	size   int
	mode   Mode
	eng    *engine.Game
	bot    *bot.Heuristic
	dead   map[engine.Point]struct{}
	result engine.Result

	hover     *engine.Point
	flash     *engine.Point
	flashLeft int
	botWait   int

	course *learn.Run
	tsume  *tsumego.Run

	handicap  int
	komi      float64
	clockMin  int
	botStrong bool
	clockLeft [3]time.Duration
	clockAt   time.Time
	showMoves bool
	saveMsg   string
	loadFiles []string
	loadErr   string

	menuSizeBtns []button
	menuModeBtns []button
	menuHaBtns   []button
	menuKomiBtns []button
	menuClkBtns  []button
	menuBotBtns  []button
	startBtn     button
	learnBtn     button
	tsumeBtn     button
	loadBtn      button
	playBtns     []button
	scoreBtns    []button
	resultBtn    button
	learnNav     []button
	tsumeNav     []button
	loadNav      []button
	quitBtn      button
	setupPlay    button
	setupBack    button
	wantQuit     bool
}

func NewApp() *App {
	a := &App{
		size: 9,
		mode: HumanHuman,
		komi: engine.DefaultKomi,
		bot:  bot.New(rand.New(rand.NewSource(time.Now().UnixNano()))),
		dead: map[engine.Point]struct{}{},
	}
	a.layoutChrome()
	return a
}

func (a *App) layoutChrome() {
	const bw, bh, gap float32 = 200, 40, 12
	row := func(y float32, labels []struct{ lab, id string }) []button {
		out := make([]button, len(labels))
		n := float32(len(labels))
		w := bw
		if n > 4 {
			w = 92
		}
		total := n*w + (n-1)*gap
		x := (float32(screenW) - total) / 2
		for i, it := range labels {
			out[i] = button{x + float32(i)*(w+gap), y, w, bh, it.lab, it.id}
		}
		return out
	}
	homeW, homeH := float32(320), float32(52)
	homeX := (float32(screenW) - homeW) / 2
	a.startBtn = button{homeX, 280, homeW, homeH, "Start game", "start"}
	a.learnBtn = button{homeX, 348, homeW, homeH, "Learn to play", "learn"}
	a.tsumeBtn = button{homeX, 416, homeW, homeH, "Tsumego", "tsume"}
	a.loadBtn = button{homeX, 484, homeW, homeH, "Load SGF", "load"}
	a.quitBtn = button{homeX, 552, homeW, homeH, "Quit", "quit"}

	a.menuSizeBtns = row(200, []struct{ lab, id string }{{"9 × 9", "9"}, {"13 × 13", "13"}, {"19 × 19", "19"}})
	a.menuModeBtns = row(280, []struct{ lab, id string }{{"Humans", "hh"}, {"Black vs bot", "hb"}, {"White vs bot", "hw"}})
	a.menuBotBtns = row(360, []struct{ lab, id string }{{"Bot easy", "be"}, {"Bot club", "bc"}})
	a.menuHaBtns = row(440, []struct{ lab, id string }{{"None", "h0"}, {"2", "h2"}, {"3", "h3"}, {"4", "h4"}, {"5", "h5"}, {"9", "h9"}})
	a.menuKomiBtns = row(520, []struct{ lab, id string }{{"komi 0", "k0"}, {"0.5", "k05"}, {"6.5", "k65"}, {"7.5", "k75"}})
	a.menuClkBtns = row(600, []struct{ lab, id string }{{"No clock", "c0"}, {"5 min", "c5"}, {"10 min", "c10"}, {"15 min", "c15"}})
	a.setupPlay = button{float32(screenW)/2 - 200, 700, 180, 48, "Play", "play"}
	a.setupBack = button{float32(screenW)/2 + 20, 700, 180, 48, "Back", "back"}

	learnY := float32(screenH - learnBotH + 28)
	a.learnNav = []button{
		{40, learnY, 130, 44, "Back", "back"},
		{180, learnY, 130, 44, "Skip", "skip"},
		{320, learnY, 140, 44, "Pass (P)", "pass"},
		{470, learnY, 180, 44, "Next lesson", "next"},
		{660, learnY, 140, 44, "Play 9×9", "play9"},
		{810, learnY, 120, 44, "Menu", "menu"},
	}

	hudY := float32(screenH - hudH + 18)
	a.playBtns = []button{
		{24, hudY, 100, 38, "Pass (P)", "pass"},
		{132, hudY, 100, 38, "Undo (U)", "undo"},
		{240, hudY, 110, 38, "Resign (R)", "resign"},
		{358, hudY, 100, 38, "Save (S)", "save"},
		{466, hudY, 120, 38, "Moves (M)", "moves"},
		{836, hudY, 100, 38, "Menu", "menu"},
	}
	a.scoreBtns = []button{
		{40, hudY, 180, 44, "Confirm score", "confirm"},
		{240, hudY, 140, 44, "Resume", "resume"},
		{780, hudY, 140, 44, "Menu", "menu"},
	}
	a.resultBtn = button{float32(screenW)/2 - 120, float32(screenH - 88), 240, 44, "New game", "new"}

	ty := float32(screenH - learnBotH + 28)
	a.tsumeNav = []button{
		{40, ty, 130, 44, "Back", "back"},
		{180, ty, 130, 44, "Reset", "reset"},
		{320, ty, 160, 44, "Next problem", "next"},
		{810, ty, 120, 44, "Menu", "menu"},
	}
	a.loadNav = []button{
		{40, ty, 130, 44, "Refresh", "refresh"},
		{810, ty, 120, 44, "Menu", "menu"},
	}
}

func (a *App) Layout(int, int) (int, int) { return screenW, screenH }

func (a *App) Update() error {
	applyChrome(theme.Tick())
	if a.wantQuit {
		return ebiten.Termination
	}
	mx, my := ebiten.CursorPosition()
	clicked, cx, cy := justClick()

	switch a.scene {
	case sceneMenu:
		a.updateMenu(clicked, cx, cy)
	case sceneSetup:
		a.updateSetup(clicked, cx, cy)
	case scenePlay:
		a.updatePlay(clicked, cx, cy, mx, my)
	case sceneScore:
		a.updateScore(clicked, cx, cy)
	case sceneResult:
		if clicked && a.resultBtn.contains(cx, cy) {
			a.scene = sceneSetup
		}
	case sceneLearn:
		a.updateLearn(clicked, cx, cy, mx, my)
	case sceneTsumego:
		a.updateTsumego(clicked, cx, cy, mx, my)
	case sceneLoad:
		a.updateLoad(clicked, cx, cy)
	}
	a.handleDrops()
	return nil
}

func (a *App) updateMenu(clicked bool, cx, cy int) {
	if !clicked {
		return
	}
	if a.startBtn.contains(cx, cy) {
		a.scene = sceneSetup
		return
	}
	if a.learnBtn.contains(cx, cy) {
		a.startLearn()
		return
	}
	if a.tsumeBtn.contains(cx, cy) {
		a.startTsumego()
		return
	}
	if a.loadBtn.contains(cx, cy) {
		a.openLoad()
		return
	}
	if a.quitBtn.contains(cx, cy) {
		a.wantQuit = true
	}
}

func (a *App) updateSetup(clicked bool, cx, cy int) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.scene = sceneMenu
		return
	}
	if !clicked {
		return
	}
	if a.setupBack.contains(cx, cy) {
		a.scene = sceneMenu
		return
	}
	if a.setupPlay.contains(cx, cy) {
		a.startGame()
		return
	}
	for _, b := range a.menuSizeBtns {
		if b.contains(cx, cy) {
			switch b.id {
			case "9":
				a.size = 9
			case "13":
				a.size = 13
			case "19":
				a.size = 19
			}
		}
	}
	for _, b := range a.menuModeBtns {
		if b.contains(cx, cy) {
			switch b.id {
			case "hh":
				a.mode = HumanHuman
			case "hb":
				a.mode = HumanBlack
			case "hw":
				a.mode = HumanWhite
			}
		}
	}
	if a.mode.vsBot() {
		for _, b := range a.menuBotBtns {
			if b.contains(cx, cy) {
				a.botStrong = b.id == "bc"
			}
		}
	}
	for _, b := range a.menuHaBtns {
		if b.contains(cx, cy) {
			switch b.id {
			case "h0":
				a.handicap = 0
			case "h2":
				a.handicap, a.komi = 2, 0.5
			case "h3":
				a.handicap, a.komi = 3, 0.5
			case "h4":
				a.handicap, a.komi = 4, 0.5
			case "h5":
				a.handicap, a.komi = 5, 0.5
			case "h9":
				a.handicap, a.komi = 9, 0.5
			}
		}
	}
	for _, b := range a.menuKomiBtns {
		if b.contains(cx, cy) {
			switch b.id {
			case "k0":
				a.komi = 0
			case "k05":
				a.komi = 0.5
			case "k65":
				a.komi = 6.5
			case "k75":
				a.komi = 7.5
			}
		}
	}
	for _, b := range a.menuClkBtns {
		if b.contains(cx, cy) {
			switch b.id {
			case "c0":
				a.clockMin = 0
			case "c5":
				a.clockMin = 5
			case "c10":
				a.clockMin = 10
			case "c15":
				a.clockMin = 15
			}
		}
	}
}

func (a *App) startGame() {
	g, err := engine.New(a.size, a.komi)
	if err != nil {
		return
	}
	if a.handicap >= 2 {
		if err := g.PlaceHandicap(a.handicap); err != nil {
			return
		}
	}
	a.eng = g
	a.dead = map[engine.Point]struct{}{}
	a.scene = scenePlay
	a.botWait = 0
	a.showMoves = false
	a.saveMsg = ""
	a.bot.Strong = a.botStrong
	a.resetClock()
	if a.botToPlay() {
		a.botWait = 24
	}
}

func (a *App) botToPlay() bool {
	if a.eng == nil || a.eng.Over() {
		return false
	}
	switch a.mode {
	case HumanBlack:
		return a.eng.ToPlay() == engine.White
	case HumanWhite:
		return a.eng.ToPlay() == engine.Black
	default:
		return false
	}
}

func (a *App) updatePlay(clicked bool, cx, cy, mx, my int) {
	if a.flashLeft > 0 {
		a.flashLeft--
		if a.flashLeft == 0 {
			a.flash = nil
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.scene = sceneMenu
		return
	}
	if !a.botToPlay() {
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			a.doPass()
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyU) {
			a.doUndo()
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			_ = a.eng.Resign()
			a.finish()
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyS) {
			a.saveSGF()
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyM) {
			a.showMoves = !a.showMoves
			return
		}
	}

	bv := layoutBoard(screenW, screenH, hudH, a.eng.Size())
	if p, ok := bv.hit(mx, my); ok {
		cp := p
		a.hover = &cp
	} else {
		a.hover = nil
	}

	a.tickClock()

	if a.botToPlay() {
		a.hover = nil
		if a.botWait > 0 {
			a.botWait--
			return
		}
		a.botMove()
		return
	}

	if !clicked {
		return
	}
	for _, b := range a.playBtns {
		if !b.contains(cx, cy) {
			continue
		}
		switch b.id {
		case "pass":
			a.doPass()
		case "undo":
			a.doUndo()
		case "resign":
			_ = a.eng.Resign()
			a.finish()
		case "menu":
			a.scene = sceneMenu
		case "save":
			a.saveSGF()
		case "moves":
			a.showMoves = !a.showMoves
		}
		return
	}
	if p, ok := bv.hit(cx, cy); ok {
		if err := a.eng.Play(p); err != nil {
			fp := p
			a.flash = &fp
			a.flashLeft = 18
			return
		}
		if a.eng.Over() {
			a.enterScore()
			return
		}
		if a.botToPlay() {
			a.botWait = 18
		}
	}
}

func (a *App) botMove() {
	p, ok := a.bot.Pick(a.eng)
	if !ok {
		a.doPass()
		return
	}
	if err := a.eng.Play(p); err != nil {
		a.doPass()
		return
	}
	if a.eng.Over() {
		a.enterScore()
	}
}

func (a *App) doPass() {
	if err := a.eng.Pass(); err != nil {
		return
	}
	if a.eng.Over() {
		a.enterScore()
		return
	}
	if a.botToPlay() {
		a.botWait = 18
	}
}

func (a *App) doUndo() {
	_ = a.eng.Undo()
	if a.mode != HumanHuman && a.botToPlay() {
		_ = a.eng.Undo()
	}
	a.botWait = 0
}

func (a *App) enterScore() {
	if a.eng.Resigned() != engine.Empty {
		a.finish()
		return
	}
	a.dead = map[engine.Point]struct{}{}
	a.scene = sceneScore
}

func (a *App) finish() {
	a.result = a.eng.Score(a.dead)
	a.scene = sceneResult
}

func (a *App) updateScore(clicked bool, cx, cy int) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.scene = sceneMenu
		return
	}
	if !clicked {
		return
	}
	for _, b := range a.scoreBtns {
		if !b.contains(cx, cy) {
			continue
		}
		switch b.id {
		case "confirm":
			a.finish()
		case "resume":
			if err := a.eng.ResumeAfterPasses(); err == nil {
				a.scene = scenePlay
			}
		case "menu":
			a.scene = sceneMenu
		}
		return
	}
	bv := layoutBoard(screenW, screenH, hudH, a.eng.Size())
	p, ok := bv.hit(cx, cy)
	if !ok {
		return
	}
	c := a.eng.At(p)
	if c == engine.Empty {
		return
	}
	grp := a.eng.Group(p)
	_, marked := a.dead[p]
	for _, q := range grp {
		if marked {
			delete(a.dead, q)
		} else {
			a.dead[q] = struct{}{}
		}
	}
}

func (a *App) Draw(screen *ebiten.Image) {
	applyChrome(theme.Current())
	screen.Fill(colHUD)
	mx, my := ebiten.CursorPosition()
	switch a.scene {
	case sceneMenu:
		a.drawMenu(screen, mx, my)
	case sceneSetup:
		a.drawSetup(screen, mx, my)
	case scenePlay:
		a.drawPlay(screen, mx, my)
	case sceneScore:
		a.drawScore(screen, mx, my)
	case sceneResult:
		a.drawResult(screen, mx, my)
	case sceneLearn:
		a.drawLearn(screen, mx, my)
	case sceneTsumego:
		a.drawTsumego(screen, mx, my)
	case sceneLoad:
		a.drawLoad(screen, mx, my)
	}
	drawLabel(screen, meta.Copyright, screenW/2, float32(screenH-20), faceSm, colMuted, true)
}

func (a *App) drawMenu(screen *ebiten.Image, mx, my int) {
	drawLabel(screen, "Go with Go", screenW/2, 140, faceBig, colInk, true)
	drawLabel(screen, "Japanese rules  ·  simple ko", screenW/2, 188, face, colMuted, true)
	drawButton(screen, a.startBtn, mx, my)
	drawButton(screen, a.learnBtn, mx, my)
	drawButton(screen, a.tsumeBtn, mx, my)
	drawButton(screen, a.loadBtn, mx, my)
	drawButton(screen, a.quitBtn, mx, my)
}

func (a *App) drawSetup(screen *ebiten.Image, mx, my int) {
	drawLabel(screen, "New game", screenW/2, 80, faceBig, colInk, true)
	drawLabel(screen, "Board", screenW/2, 168, faceSm, colMuted, true)
	for _, b := range a.menuSizeBtns {
		drawButton(screen, b, mx, my)
		if (b.id == "9" && a.size == 9) || (b.id == "13" && a.size == 13) || (b.id == "19" && a.size == 19) {
			vectorStrokeSelected(screen, b)
		}
	}
	drawLabel(screen, "Players  ·  Black vs bot means you play Black", screenW/2, 248, faceSm, colMuted, true)
	for _, b := range a.menuModeBtns {
		drawButton(screen, b, mx, my)
		if (b.id == "hh" && a.mode == HumanHuman) || (b.id == "hb" && a.mode == HumanBlack) || (b.id == "hw" && a.mode == HumanWhite) {
			vectorStrokeSelected(screen, b)
		}
	}
	if a.mode.vsBot() {
		drawLabel(screen, "Bot", screenW/2, 328, faceSm, colMuted, true)
		for _, b := range a.menuBotBtns {
			drawButton(screen, b, mx, my)
			if (b.id == "bc" && a.botStrong) || (b.id == "be" && !a.botStrong) {
				vectorStrokeSelected(screen, b)
			}
		}
	}
	drawLabel(screen, "Handicap  ·  2+ stones for Black, White moves first, komi becomes 0.5", screenW/2, 408, faceSm, colMuted, true)
	for _, b := range a.menuHaBtns {
		drawButton(screen, b, mx, my)
		want := map[string]int{"h0": 0, "h2": 2, "h3": 3, "h4": 4, "h5": 5, "h9": 9}[b.id]
		if a.handicap == want {
			vectorStrokeSelected(screen, b)
		}
	}
	drawLabel(screen, "Komi (added to White)", screenW/2, 488, faceSm, colMuted, true)
	for _, b := range a.menuKomiBtns {
		drawButton(screen, b, mx, my)
		want := map[string]float64{"k0": 0, "k05": 0.5, "k65": 6.5, "k75": 7.5}[b.id]
		if a.komi == want {
			vectorStrokeSelected(screen, b)
		}
	}
	drawLabel(screen, "Clock (each player)", screenW/2, 568, faceSm, colMuted, true)
	for _, b := range a.menuClkBtns {
		drawButton(screen, b, mx, my)
		want := map[string]int{"c0": 0, "c5": 5, "c10": 10, "c15": 15}[b.id]
		if a.clockMin == want {
			vectorStrokeSelected(screen, b)
		}
	}
	drawButton(screen, a.setupPlay, mx, my)
	drawButton(screen, a.setupBack, mx, my)
}

func vectorStrokeSelected(screen *ebiten.Image, b button) {
	vectorStrokeRect(screen, b.x-3, b.y-3, b.w+6, b.h+6, colWin)
}

func vectorStrokeRect(dst *ebiten.Image, x, y, w, h float32, col color.Color) {
	vector.StrokeLine(dst, x, y, x+w, y, 2, col, true)
	vector.StrokeLine(dst, x+w, y, x+w, y+h, 2, col, true)
	vector.StrokeLine(dst, x+w, y+h, x, y+h, 2, col, true)
	vector.StrokeLine(dst, x, y+h, x, y, 2, col, true)
}

func (a *App) drawPlay(screen *ebiten.Image, mx, my int) {
	bv := layoutBoard(screenW, screenH, hudH, a.eng.Size())
	bv.draw(screen, a.eng, a.hover, nil, a.flash, nil, nil)
	drawPanel(screen, 0, float32(screenH-hudH), screenW, hudH)
	status := fmt.Sprintf("%s to play", a.eng.ToPlay())
	if a.botToPlay() {
		status += "  ·  bot"
	}
	drawLabel(screen, status, 24, float32(screenH-hudH+64), faceSm, colInk, false)
	caps := fmt.Sprintf("Prisoners B %d  W %d   komi %.1f   %s",
		a.eng.Captured(engine.Black), a.eng.Captured(engine.White), a.eng.Komi(), a.clockText())
	drawLabel(screen, caps, 24, float32(screenH-hudH+86), faceSm, colMuted, false)
	if lm := a.eng.LastMove(); lm != nil {
		note := lm.Label(a.eng.Size())
		if lm.Comment != "" && lm.Captured == 0 && !lm.Pass && !lm.Resign {
			note += " · " + lm.Comment
		}
		drawLabel(screen, fmt.Sprintf("Move %d  %s", a.eng.MoveCount(), note), 24, float32(screenH-hudH+108), faceSm, colInk, false)
	}
	if a.saveMsg != "" {
		drawLabel(screen, a.saveMsg, 520, float32(screenH-hudH+108), faceSm, colWin, false)
	}
	for _, b := range a.playBtns {
		drawButton(screen, b, mx, my)
	}
	if a.showMoves {
		a.drawMoveList(screen)
	}
}

func (a *App) drawScore(screen *ebiten.Image, mx, my int) {
	bv := layoutBoard(screenW, screenH, hudH, a.eng.Size())
	bv.draw(screen, a.eng, nil, a.dead, nil, nil, a.eng.Territory(a.dead))
	drawPanel(screen, 0, float32(screenH-hudH), screenW, hudH)
	preview := a.eng.Score(a.dead)
	drawLabel(screen, "Click a group to mark it dead. Shaded empty points are territory.", 24, float32(screenH-hudH+62), faceSm, colMuted, false)
	drawLabel(screen, "Dark = Black land   Light = White land   Small brown = dame (nobody)", 24, float32(screenH-hudH+84), faceSm, colMuted, false)
	sum := fmt.Sprintf("Preview  Black %.1f (land %d + prisoners %d)    White %.1f (land %d + prisoners %d + komi %.1f)",
		preview.Black, preview.BlackTerritory, preview.BlackCaptures,
		preview.White, preview.WhiteTerritory, preview.WhiteCaptures, preview.Komi)
	drawLabel(screen, sum, 24, float32(screenH-hudH+108), faceSm, colInk, false)
	for _, b := range a.scoreBtns {
		drawButton(screen, b, mx, my)
	}
}

func (a *App) drawResult(screen *ebiten.Image, mx, my int) {
	const resultH = 260
	if a.eng != nil {
		bv := layoutBoard(screenW, screenH, resultH, a.eng.Size())
		var terr map[engine.Point]engine.Color
		if a.result.Resigned == engine.Empty {
			terr = a.eng.Territory(a.dead)
		}
		bv.draw(screen, a.eng, nil, a.dead, nil, nil, terr)
	}
	drawPanel(screen, 0, float32(screenH-resultH), screenW, resultH)
	r := a.result
	top := float32(screenH - resultH)
	drawLabel(screen, r.String(), screenW/2, top+12, faceBig, colWin, true)
	if r.Resigned == engine.Empty {
		drawLabel(screen, "Dark marks = Black land    Light marks = White land    Small brown = dame (nobody)", screenW/2, top+52, faceSm, colMuted, true)
		drawLabel(screen, fmt.Sprintf("Black  %d land + %d prisoners = %.1f", r.BlackTerritory, r.BlackCaptures, r.Black), screenW/2, top+80, face, colInk, true)
		drawLabel(screen, fmt.Sprintf("White  %d land + %d prisoners + %.1f komi = %.1f", r.WhiteTerritory, r.WhiteCaptures, r.Komi, r.White), screenW/2, top+106, face, colInk, true)
		drawLabel(screen, "Stones still on the board do not count.", screenW/2, top+134, faceSm, colMuted, true)
	}
	drawButton(screen, a.resultBtn, mx, my)
}

func justClick() (bool, int, int) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return true, x, y
	}
	ids := inpututil.AppendJustPressedTouchIDs(nil)
	if len(ids) > 0 {
		x, y := ebiten.TouchPosition(ids[0])
		return true, x, y
	}
	return false, 0, 0
}
