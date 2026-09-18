package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"go-with-go/engine"
	"go-with-go/learn"
	"go-with-go/tsumego"
)

func (m model) startLearn() (tea.Model, tea.Cmd) {
	r, err := learn.NewRun()
	if err != nil {
		m.status = err.Error()
		return m, nil
	}
	m.course = r
	m.eng = r.Game()
	m.size = r.Game().Size()
	m.cursor = engine.Point{X: m.size / 2, Y: m.size / 2}
	m.scene = sceneLearn
	m.status = "enter play  ·  p pass  ·  n next  ·  [ ] skip/back  ·  q menu"
	return m, nil
}

func (m model) learnKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.course == nil {
		m.scene = sceneMenu
		return m, nil
	}
	switch msg.String() {
	case "left", "h":
		if m.cursor.X > 0 {
			m.cursor.X--
		}
	case "right", "l":
		if m.cursor.X < m.size-1 {
			m.cursor.X++
		}
	case "up", "k":
		if m.cursor.Y > 0 {
			m.cursor.Y--
		}
	case "down", "j":
		if m.cursor.Y < m.size-1 {
			m.cursor.Y++
		}
	case "enter", " ":
		return m.learnClick(m.cursor)
	case "p":
		m.course.Pass()
	case "n", "]":
		return m.learnNext()
	case "[":
		_ = m.course.PrevLesson()
		if g := m.course.Game(); g != nil {
			m.eng = g
			m.size = g.Size()
		}
	case "esc", "q":
		m.scene = sceneMenu
	}
	return m, nil
}

func (m model) learnClick(p engine.Point) (tea.Model, tea.Cmd) {
	if m.course == nil {
		return m, nil
	}
	m.course.Click(p)
	return m, nil
}

func (m model) learnNext() (tea.Model, tea.Cmd) {
	if m.course.Graduated() {
		m.size = 9
		m.mode = humanBlack
		return m.start()
	}
	_ = m.course.NextLesson()
	if g := m.course.Game(); g != nil {
		m.eng = g
		m.size = g.Size()
	}
	return m, nil
}

func (m model) startTsumego() (tea.Model, tea.Cmd) {
	r, err := tsumego.NewRun()
	if err != nil {
		m.status = err.Error()
		return m, nil
	}
	m.tsume = r
	m.eng = r.Game()
	m.size = r.Game().Size()
	m.cursor = engine.Point{X: m.size / 2, Y: m.size / 2}
	m.scene = sceneTsumego
	m.status = "enter play  ·  n next  ·  r reset  ·  q menu"
	return m, nil
}

func (m model) tsumeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.tsume == nil {
		m.scene = sceneMenu
		return m, nil
	}
	switch msg.String() {
	case "left", "h":
		if m.cursor.X > 0 {
			m.cursor.X--
		}
	case "right", "l":
		if m.cursor.X < m.size-1 {
			m.cursor.X++
		}
	case "up", "k":
		if m.cursor.Y > 0 {
			m.cursor.Y--
		}
	case "down", "j":
		if m.cursor.Y < m.size-1 {
			m.cursor.Y++
		}
	case "enter", " ":
		return m.tsumeClick(m.cursor)
	case "n":
		_ = m.tsume.Next()
		m.eng = m.tsume.Game()
	case "r":
		_ = m.tsume.Reset()
		m.eng = m.tsume.Game()
	case "[":
		_ = m.tsume.Prev()
		m.eng = m.tsume.Game()
	case "esc", "q":
		m.scene = sceneMenu
	}
	return m, nil
}

func (m model) tsumeClick(p engine.Point) (tea.Model, tea.Cmd) {
	if m.tsume == nil {
		return m, nil
	}
	m.tsume.Click(p)
	return m, nil
}

func tuiSaveSGF(g *engine.Game) string {
	if g == nil {
		return "no game"
	}
	if err := os.MkdirAll("games", 0o755); err != nil {
		return err.Error()
	}
	name := time.Now().Format("20060102-150405") + ".sgf"
	path := filepath.Join("games", name)
	if err := os.WriteFile(path, []byte(g.SGF()+"\n"), 0o644); err != nil {
		return err.Error()
	}
	return fmt.Sprintf("saved %s", path)
}
