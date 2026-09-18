package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"go-with-go/learn"
)

const (
	learnTopH = 176
	learnBotH = 100
)

func (a *App) startLearn() {
	r, err := learn.NewRun()
	if err != nil {
		return
	}
	a.course = r
	a.eng = r.Game()
	a.hover = nil
	a.flash = nil
	a.scene = sceneLearn
}

func (a *App) learnBoard() boardView {
	return layoutBoardPad(screenW, screenH, learnTopH, learnBotH, a.eng.Size())
}

func (a *App) updateLearn(clicked bool, cx, cy, mx, my int) {
	if a.course == nil {
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
	if a.eng != nil {
		bv := a.learnBoard()
		if p, ok := bv.hit(mx, my); ok {
			cp := p
			a.hover = &cp
		} else {
			a.hover = nil
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		a.course.Pass()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) && a.course.Complete() {
		a.learnNext()
		return
	}

	if !clicked {
		return
	}
	for _, b := range a.learnNav {
		if !b.contains(cx, cy) {
			continue
		}
		switch b.id {
		case "back":
			_ = a.course.PrevLesson()
			a.eng = a.course.Game()
		case "skip", "next":
			a.learnNext()
		case "pass":
			a.course.Pass()
		case "menu":
			a.scene = sceneMenu
		case "play9":
			a.size = 9
			a.mode = HumanBlack
			a.startGame()
		}
		return
	}
	if a.eng == nil || a.course.Complete() || a.course.Graduated() {
		return
	}
	bv := a.learnBoard()
	p, ok := bv.hit(cx, cy)
	if !ok {
		return
	}
	if a.course.Click(p) {
		fp := p
		a.flash = &fp
		a.flashLeft = 18
	}
}

func (a *App) learnNext() {
	if a.course.Graduated() {
		a.size = 9
		a.mode = HumanBlack
		a.startGame()
		return
	}
	_ = a.course.NextLesson()
	if a.course.Game() != nil {
		a.eng = a.course.Game()
	}
}

func (a *App) drawLearn(screen *ebiten.Image, mx, my int) {
	if a.course == nil {
		return
	}
	les := a.course.Lesson()
	drawPanel(screen, 0, 0, screenW, learnTopH)
	n := a.course.Index() + 1
	if a.course.Graduated() {
		n = a.course.Count()
	}
	head := fmt.Sprintf("Learn  ·  %d / %d  ·  %s", n, a.course.Count(), les.Title)
	drawLabel(screen, head, 40, 18, face, colWin, false)
	drawWrapped(screen, les.Blurb, 40, 48, screenW-80, 20, faceSm, colMuted)
	promptCol := colInk
	if a.course.Hint() != "" && !a.course.Complete() {
		promptCol = colWin
	}
	drawWrapped(screen, a.course.Prompt(), 40, 96, screenW-80, 22, face, promptCol)

	if a.eng != nil {
		bv := a.learnBoard()
		bv.draw(screen, a.eng, a.hover, nil, a.flash, a.course.Marks(), nil)
	}

	drawPanel(screen, 0, float32(screenH-learnBotH), screenW, learnBotH)
	for _, b := range a.learnNav {
		switch b.id {
		case "next":
			if !a.course.Complete() || a.course.Graduated() {
				continue
			}
		case "play9":
			if !a.course.Graduated() {
				continue
			}
		case "skip":
			if a.course.Complete() || a.course.Graduated() {
				continue
			}
		case "pass":
			if a.course.Complete() || a.course.Graduated() {
				continue
			}
		}
		drawButton(screen, b, mx, my)
	}
}
