package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"go-with-go/tsumego"
)

func (a *App) startTsumego() {
	r, err := tsumego.NewRun()
	if err != nil {
		return
	}
	a.tsume = r
	a.eng = r.Game()
	a.hover = nil
	a.flash = nil
	a.scene = sceneTsumego
}

func (a *App) updateTsumego(clicked bool, cx, cy, mx, my int) {
	if a.tsume == nil {
		a.scene = sceneMenu
		return
	}
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
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		_ = a.tsume.Next()
		a.eng = a.tsume.Game()
		return
	}
	if a.eng != nil {
		bv := layoutBoardPad(screenW, screenH, learnTopH, learnBotH, a.eng.Size())
		if p, ok := bv.hit(mx, my); ok {
			cp := p
			a.hover = &cp
		} else {
			a.hover = nil
		}
	}
	if !clicked {
		return
	}
	for _, b := range a.tsumeNav {
		if !b.contains(cx, cy) {
			continue
		}
		switch b.id {
		case "back":
			_ = a.tsume.Prev()
			a.eng = a.tsume.Game()
		case "reset":
			_ = a.tsume.Reset()
			a.eng = a.tsume.Game()
		case "next":
			_ = a.tsume.Next()
			a.eng = a.tsume.Game()
		case "menu":
			a.scene = sceneMenu
		}
		return
	}
	if a.eng == nil || a.tsume.Solved() {
		return
	}
	bv := layoutBoardPad(screenW, screenH, learnTopH, learnBotH, a.eng.Size())
	p, ok := bv.hit(cx, cy)
	if !ok {
		return
	}
	if a.tsume.Click(p) {
		fp := p
		a.flash = &fp
		a.flashLeft = 18
	}
}

func (a *App) drawTsumego(screen *ebiten.Image, mx, my int) {
	if a.tsume == nil {
		return
	}
	pr := a.tsume.Problem()
	drawPanel(screen, 0, 0, screenW, learnTopH)
	head := fmt.Sprintf("Tsumego  ·  %d / %d  ·  %s  ·  %s", a.tsume.Index()+1, a.tsume.Count(), pr.Rank, pr.Title)
	drawLabel(screen, head, 40, 18, face, colWin, false)
	col := colInk
	if a.tsume.Solved() {
		col = colWin
	}
	drawWrapped(screen, a.tsume.Prompt(), 40, 56, screenW-80, 22, face, col)

	if a.eng != nil {
		bv := layoutBoardPad(screenW, screenH, learnTopH, learnBotH, a.eng.Size())
		bv.draw(screen, a.eng, a.hover, nil, a.flash, a.tsume.Marks(), nil)
	}
	drawPanel(screen, 0, float32(screenH-learnBotH), screenW, learnBotH)
	for _, b := range a.tsumeNav {
		drawButton(screen, b, mx, my)
	}
}
