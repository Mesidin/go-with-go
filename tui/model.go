package tui

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"go-with-go/bot"
	"go-with-go/engine"
	"go-with-go/learn"
	"go-with-go/tsumego"
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
)

type mode int

const (
	humanHuman mode = iota
	humanBlack
	humanWhite
)

type model struct {
	scene  scene
	size   int
	mode   mode
	sizeI  int
	modeI  int
	eng    *engine.Game
	bot    *bot.Heuristic
	dead   map[engine.Point]struct{}
	cursor engine.Point
	result engine.Result
	status string
	width  int
	height int
	course    *learn.Run
	tsume     *tsumego.Run
	botStrong bool
}

func New() tea.Model {
	return model{
		size:  9,
		mode:  humanHuman,
		bot:   bot.New(rand.New(rand.NewSource(time.Now().UnixNano()))),
		dead:  map[engine.Point]struct{}{},
		status: "arrows / hjkl move  ·  enter play  ·  p pass  ·  u undo  ·  r resign  ·  q quit",
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.scene == scenePlay || m.scene == sceneScore || m.scene == sceneResult || m.scene == sceneLearn || m.scene == sceneTsumego {
			if msg.String() == "q" && m.scene != sceneMenu {
				m.scene = sceneMenu
				return m, nil
			}
		}
		if msg.String() == "ctrl+c" || m.scene == sceneMenu {
			return m, tea.Quit
		}
	}

	switch m.scene {
	case sceneMenu:
		return m.menuKey(msg)
	case sceneSetup:
		return m.setupKey(msg)
	case scenePlay:
		return m.playKey(msg)
	case sceneScore:
		return m.scoreKey(msg)
	case sceneResult:
		if msg.String() == "enter" || msg.String() == "n" {
			m.scene = sceneMenu
		}
	case sceneLearn:
		return m.learnKey(msg)
	case sceneTsumego:
		return m.tsumeKey(msg)
	}
	return m, nil
}

func (m model) menuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.scene = sceneSetup
		m.status = "enter play  ·  esc back"
		return m, nil
	case "t":
		return m.startLearn()
	case "g":
		return m.startTsumego()
	case "l":
		m.status = "drop an .sgf on the desktop game, or put files in games/"
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) setupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	sizes := []int{9, 13, 19}
	modes := []mode{humanHuman, humanBlack, humanWhite}
	switch msg.String() {
	case "left", "h":
		if m.sizeI > 0 {
			m.sizeI--
		}
		m.size = sizes[m.sizeI]
	case "right", "l":
		if m.sizeI < 2 {
			m.sizeI++
		}
		m.size = sizes[m.sizeI]
	case "up", "k":
		if m.modeI > 0 {
			m.modeI--
		}
		m.mode = modes[m.modeI]
	case "down", "j":
		if m.modeI < 2 {
			m.modeI++
		}
		m.mode = modes[m.modeI]
	case "b":
		if m.mode != humanHuman {
			m.botStrong = !m.botStrong
		}
	case "enter":
		return m.start()
	case "esc":
		m.scene = sceneMenu
	case "q":
		m.scene = sceneMenu
	}
	return m, nil
}

func (m model) start() (tea.Model, tea.Cmd) {
	g, err := engine.New(m.size, engine.DefaultKomi)
	if err != nil {
		m.status = err.Error()
		return m, nil
	}
	m.eng = g
	m.dead = map[engine.Point]struct{}{}
	m.cursor = engine.Point{X: m.size / 2, Y: m.size / 2}
	m.scene = scenePlay
	m.bot.Strong = m.botStrong
	m.status = "enter place  ·  p pass  ·  u undo  ·  r resign  ·  q menu"
	if m.botToPlay() {
		return m.botMove()
	}
	return m, nil
}

func (m model) botToPlay() bool {
	if m.eng == nil || m.eng.Over() {
		return false
	}
	switch m.mode {
	case humanBlack:
		return m.eng.ToPlay() == engine.White
	case humanWhite:
		return m.eng.ToPlay() == engine.Black
	default:
		return false
	}
}

func (m model) playKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.eng == nil {
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
		return m.tryPlay(m.cursor)
	case "p":
		return m.doPass()
	case "s":
		m.status = tuiSaveSGF(m.eng)
		return m, nil
	case "u":
		return m.doUndo()
	case "r":
		_ = m.eng.Resign()
		return m.finish()
	case "esc", "q":
		m.scene = sceneMenu
	}
	return m, nil
}

func (m model) tryPlay(p engine.Point) (tea.Model, tea.Cmd) {
	if err := m.eng.Play(p); err != nil {
		m.status = err.Error()
		return m, nil
	}
	m.status = ""
	if m.eng.Over() {
		return m.enterScore()
	}
	if m.botToPlay() {
		return m.botMove()
	}
	return m, nil
}

func (m model) doPass() (tea.Model, tea.Cmd) {
	if err := m.eng.Pass(); err != nil {
		m.status = err.Error()
		return m, nil
	}
	if m.eng.Over() {
		return m.enterScore()
	}
	if m.botToPlay() {
		return m.botMove()
	}
	return m, nil
}

func (m model) doUndo() (tea.Model, tea.Cmd) {
	_ = m.eng.Undo()
	if m.mode != humanHuman && m.botToPlay() {
		_ = m.eng.Undo()
	}
	m.status = "undone"
	return m, nil
}

func (m model) botMove() (tea.Model, tea.Cmd) {
	p, ok := m.bot.Pick(m.eng)
	if !ok {
		return m.doPass()
	}
	if err := m.eng.Play(p); err != nil {
		return m.doPass()
	}
	m.status = fmt.Sprintf("bot played %s", p)
	if m.eng.Over() {
		return m.enterScore()
	}
	return m, nil
}

func (m model) enterScore() (tea.Model, tea.Cmd) {
	if m.eng.Resigned() != engine.Empty {
		return m.finish()
	}
	m.dead = map[engine.Point]struct{}{}
	m.scene = sceneScore
	m.status = "enter marks dead  ·  c confirm  ·  x resume"
	return m, nil
}

func (m model) finish() (tea.Model, tea.Cmd) {
	m.result = m.eng.Score(m.dead)
	m.scene = sceneResult
	m.status = "enter / n  new game  ·  q menu"
	return m, nil
}

func (m model) scoreKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		m.toggleDead(m.cursor)
	case "c":
		return m.finish()
	case "x":
		if err := m.eng.ResumeAfterPasses(); err == nil {
			m.scene = scenePlay
			m.status = "resumed"
		}
	case "esc", "q":
		m.scene = sceneMenu
	}
	return m, nil
}

func (m model) toggleDead(p engine.Point) {
	c := m.eng.At(p)
	if c == engine.Empty {
		return
	}
	grp := m.eng.Group(p)
	_, marked := m.dead[p]
	for _, q := range grp {
		if marked {
			delete(m.dead, q)
		} else {
			m.dead[q] = struct{}{}
		}
	}
}

func (m model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	if m.scene != scenePlay && m.scene != sceneScore && m.scene != sceneLearn && m.scene != sceneTsumego {
		return m, nil
	}
	// Board is drawn starting at row 3, col 4 (after rank label).
	x := msg.X - 4
	y := msg.Y - 3
	if x < 0 || y < 0 || x >= m.size || y >= m.size {
		return m, nil
	}
	m.cursor = engine.Point{X: x, Y: y}
	if m.scene == scenePlay {
		return m.tryPlay(m.cursor)
	}
	if m.scene == sceneLearn {
		return m.learnClick(m.cursor)
	}
	if m.scene == sceneTsumego {
		return m.tsumeClick(m.cursor)
	}
	m.toggleDead(m.cursor)
	return m, nil
}
