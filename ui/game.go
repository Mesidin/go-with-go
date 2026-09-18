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
	"go-with-go/learn"
)

const (
	screenW = 960
	screenH = 1040
	hudH    = 140

	resultCardW = 560
	resultCardH = 430
)

func resultCardRect() (x, y, w, h float32) {
	return (screenW - resultCardW) / 2, (screenH - resultCardH) / 2, resultCardW, resultCardH
}

type scene int

const (
	sceneMenu scene = iota
	scenePlay
	sceneScore
	sceneResult
	sceneLearn
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
		return "You play Black"
	case HumanWhite:
		return "You play White"
	default:
		return "Two humans"
	}
}

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

	menuSizeBtns []button
	menuModeBtns []button
	startBtn     button
	learnBtn     button
	playBtns     []button
	scoreBtns    []button
	resultBtn    button
	learnNav     []button
}

func NewApp() *App {
	a := &App{
		size: 9,
		mode: HumanHuman,
		bot:  bot.New(rand.New(rand.NewSource(time.Now().UnixNano()))),
		dead: map[engine.Point]struct{}{},
	}
	a.layoutChrome()
	return a
}

func (a *App) layoutChrome() {
	const bw, bh, gap = 160, 44, 16
	startX := float32(screenW)/2 - (3*bw+2*gap)/2
	y := float32(420)
	a.menuSizeBtns = []button{
		{startX, y, bw, bh, "9 × 9", "9"},
		{startX + bw + gap, y, bw, bh, "13 × 13", "13"},
		{startX + 2*(bw+gap), y, bw, bh, "19 × 19", "19"},
	}
	y = 520
	a.menuModeBtns = []button{
		{startX, y, bw, bh, "Two humans", "hh"},
		{startX + bw + gap, y, bw, bh, "Play Black", "hb"},
		{startX + 2*(bw+gap), y, bw, bh, "Play White", "hw"},
	}
	a.startBtn = button{float32(screenW)/2 - 120, 620, 240, 52, "Start game", "start"}
	a.learnBtn = button{float32(screenW)/2 - 120, 688, 240, 52, "Learn to play", "learn"}

	learnY := float32(screenH - learnBotH + 28)
	a.learnNav = []button{
		{40, learnY, 130, 44, "Back", "back"},
		{180, learnY, 130, 44, "Skip", "skip"},
		{320, learnY, 140, 44, "Pass (P)", "pass"},
		{470, learnY, 180, 44, "Next lesson", "next"},
		{660, learnY, 140, 44, "Play 9×9", "play9"},
		{810, learnY, 120, 44, "Menu", "menu"},
	}

	hudY := float32(screenH - hudH + 24)
	a.playBtns = []button{
		{40, hudY, 140, 44, "Pass (P)", "pass"},
		{200, hudY, 140, 44, "Undo (U)", "undo"},
		{360, hudY, 140, 44, "Resign (R)", "resign"},
		{780, hudY, 140, 44, "Menu", "menu"},
	}
	a.scoreBtns = []button{
		{40, hudY, 180, 44, "Confirm score", "confirm"},
		{240, hudY, 140, 44, "Resume", "resume"},
		{780, hudY, 140, 44, "Menu", "menu"},
	}
	cx, cy, cw, ch := resultCardRect()
	a.resultBtn = button{cx + cw/2 - 120, cy + ch - 76, 240, 48, "New game", "new"}
}

func (a *App) Layout(int, int) (int, int) { return screenW, screenH }

func (a *App) Update() error {
	mx, my := ebiten.CursorPosition()
	clicked, cx, cy := justClick()

	switch a.scene {
	case sceneMenu:
		a.updateMenu(clicked, cx, cy)
	case scenePlay:
		a.updatePlay(clicked, cx, cy, mx, my)
	case sceneScore:
		a.updateScore(clicked, cx, cy)
	case sceneResult:
		if clicked && a.resultBtn.contains(cx, cy) {
			a.scene = sceneMenu
		}
	case sceneLearn:
		a.updateLearn(clicked, cx, cy, mx, my)
	}
	return nil
}

func (a *App) updateMenu(clicked bool, cx, cy int) {
	if !clicked {
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
	if a.startBtn.contains(cx, cy) {
		a.startGame()
	}
	if a.learnBtn.contains(cx, cy) {
		a.startLearn()
	}
}

func (a *App) startGame() {
	g, err := engine.New(a.size, engine.DefaultKomi)
	if err != nil {
		return
	}
	a.eng = g
	a.dead = map[engine.Point]struct{}{}
	a.scene = scenePlay
	a.botWait = 0
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
	}

	bv := layoutBoard(screenW, screenH, hudH, a.eng.Size())
	if p, ok := bv.hit(mx, my); ok {
		cp := p
		a.hover = &cp
	} else {
		a.hover = nil
	}

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
	screen.Fill(colHUD)
	mx, my := ebiten.CursorPosition()
	switch a.scene {
	case sceneMenu:
		a.drawMenu(screen, mx, my)
	case scenePlay:
		a.drawPlay(screen, mx, my)
	case sceneScore:
		a.drawScore(screen, mx, my)
	case sceneResult:
		a.drawResult(screen, mx, my)
	case sceneLearn:
		a.drawLearn(screen, mx, my)
	}
	drawLabel(screen, meta.Copyright, screenW/2, float32(screenH-20), faceSm, colMuted, true)
}

func (a *App) drawMenu(screen *ebiten.Image, mx, my int) {
	drawLabel(screen, "Go with Go", screenW/2, 160, faceBig, colInk, true)
	drawLabel(screen, "Japanese rules  ·  6.5 komi  ·  simple ko", screenW/2, 210, face, colMuted, true)
	drawLabel(screen, "Board size", screenW/2, 380, face, colMuted, true)
	for _, b := range a.menuSizeBtns {
		drawButton(screen, b, mx, my)
		if (b.id == "9" && a.size == 9) || (b.id == "13" && a.size == 13) || (b.id == "19" && a.size == 19) {
			vectorStrokeSelected(screen, b)
		}
	}
	drawLabel(screen, "Who plays", screenW/2, 490, face, colMuted, true)
	for _, b := range a.menuModeBtns {
		drawButton(screen, b, mx, my)
		if (b.id == "hh" && a.mode == HumanHuman) || (b.id == "hb" && a.mode == HumanBlack) || (b.id == "hw" && a.mode == HumanWhite) {
			vectorStrokeSelected(screen, b)
		}
	}
	drawButton(screen, a.startBtn, mx, my)
	drawButton(screen, a.learnBtn, mx, my)
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
	bv.draw(screen, a.eng, a.hover, nil, a.flash, nil)
	drawPanel(screen, 0, float32(screenH-hudH), screenW, hudH)
	status := fmt.Sprintf("%s to play", a.eng.ToPlay())
	if a.botToPlay() {
		status += "  ·  bot"
	}
	drawLabel(screen, status, 40, float32(screenH-hudH+72), face, colInk, false)
	caps := fmt.Sprintf("Prisoners  Black %d   White %d   komi %.1f",
		a.eng.Captured(engine.Black), a.eng.Captured(engine.White), a.eng.Komi())
	drawLabel(screen, caps, 400, float32(screenH-hudH+76), faceSm, colMuted, false)
	for _, b := range a.playBtns {
		drawButton(screen, b, mx, my)
	}
}

func (a *App) drawScore(screen *ebiten.Image, mx, my int) {
	bv := layoutBoard(screenW, screenH, hudH, a.eng.Size())
	bv.draw(screen, a.eng, nil, a.dead, nil, nil)
	drawPanel(screen, 0, float32(screenH-hudH), screenW, hudH)
	preview := a.eng.Score(a.dead)
	drawLabel(screen, "Click groups to mark dead  ·  Japanese territory scoring", 40, float32(screenH-hudH+72), faceSm, colMuted, false)
	sum := fmt.Sprintf("Preview  Black %.1f   White %.1f", preview.Black, preview.White)
	drawLabel(screen, sum, 400, float32(screenH-hudH+76), faceSm, colInk, false)
	for _, b := range a.scoreBtns {
		drawButton(screen, b, mx, my)
	}
}

func (a *App) drawResult(screen *ebiten.Image, mx, my int) {
	if a.eng != nil {
		bv := layoutBoard(screenW, screenH, hudH, a.eng.Size())
		bv.draw(screen, a.eng, nil, a.dead, nil, nil)
	}
	vector.FillRect(screen, 0, 0, screenW, screenH, color.RGBA{12, 8, 4, 170}, true)

	cx, cy, cw, ch := resultCardRect()
	vector.FillRect(screen, cx, cy, cw, ch, colHUD, true)
	vectorStrokeRect(screen, cx, cy, cw, ch, colWin)

	r := a.result
	drawLabel(screen, r.String(), cx+cw/2, cy+28, faceBig, colWin, true)
	drawLabel(screen, "Japanese territory scoring", cx+cw/2, cy+72, faceSm, colMuted, true)

	if r.Resigned == engine.Empty {
		drawScoreTable(screen, cx, cy+108, cw, r)
	}

	drawButton(screen, a.resultBtn, mx, my)
}

func drawScoreTable(dst *ebiten.Image, cx, cy, cw float32, r engine.Result) {
	const rowH = 34
	labelX := cx + 36
	blackX := cx + cw - 196
	whiteX := cx + cw - 48

	colB := blackX + 28
	colW := whiteX + 28
	drawLabelRight(dst, "Black", colB, cy, faceSm, colMuted)
	drawLabelRight(dst, "White", colW, cy, faceSm, colMuted)

	type row struct {
		label, black, white string
		total               bool
	}
	rows := []row{
		{"Territory", fmt.Sprintf("%d", r.BlackTerritory), fmt.Sprintf("%d", r.WhiteTerritory), false},
		{"Prisoners", fmt.Sprintf("%d", r.BlackCaptures), fmt.Sprintf("%d", r.WhiteCaptures), false},
		{"Komi", "—", fmt.Sprintf("%.1f", r.Komi), false},
		{"Total", fmt.Sprintf("%.1f", r.Black), fmt.Sprintf("%.1f", r.White), true},
	}
	y := cy + rowH
	for _, row := range rows {
		if row.total {
			vector.StrokeLine(dst, labelX, y-8, cx+cw-36, y-8, 1.2, colMuted, true)
		}
		labelCol, numFace := colInk, face
		if row.total {
			labelCol = colWin
		}
		drawLabel(dst, row.label, labelX, y, numFace, labelCol, false)
		bCol, wCol := colInk, colInk
		if row.total {
			switch r.Winner {
			case engine.Black:
				bCol, wCol = colWin, colMuted
			case engine.White:
				bCol, wCol = colMuted, colWin
			default:
				bCol, wCol = colWin, colWin
			}
		}
		drawLabelRight(dst, row.black, colB, y, numFace, bCol)
		drawLabelRight(dst, row.white, colW, y, numFace, wCol)
		y += rowH
	}
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
