package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"go-with-go/engine"
	"go-with-go/internal/meta"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("186"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	hotStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	blackSt    = lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("186"))
	whiteSt    = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("94"))
	cursorSt   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	deadSt     = lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Bold(true)
	statusSt   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
)

func (m model) View() string {
	var body string
	switch m.scene {
	case sceneMenu:
		body = m.viewMenu()
	case scenePlay:
		body = m.viewPlay()
	case sceneScore:
		body = m.viewScore()
	case sceneResult:
		body = m.viewResult()
	case sceneLearn:
		body = m.viewLearn()
	default:
		return ""
	}
	return body + "\n\n" + mutedStyle.Render(meta.Copyright)
}

func (m model) viewMenu() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Go with Go"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Japanese rules  ·  6.5 komi  ·  simple ko"))
	b.WriteString("\n\nBoard size  (h/l)\n")
	for i, s := range []int{9, 13, 19} {
		label := fmt.Sprintf("  %d × %d", s, s)
		if i == m.sizeI {
			label = hotStyle.Render("▶ " + strings.TrimSpace(label))
		}
		b.WriteString(label + "\n")
	}
	b.WriteString("\nWho plays  (j/k)\n")
	labels := []string{"Two humans", "Play Black vs bot", "Play White vs bot"}
	for i, l := range labels {
		line := "  " + l
		if i == m.modeI {
			line = hotStyle.Render("▶ " + l)
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("enter start  ·  t learn to play  ·  q quit"))
	return b.String()
}

func (m model) viewLearn() string {
	if m.course == nil {
		return "no lesson"
	}
	les := m.course.Lesson()
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Learn %d/%d  ·  %s", m.course.Index()+1, m.course.Count(), les.Title)))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(les.Blurb))
	b.WriteString("\n")
	b.WriteString(hotStyle.Render(m.course.Prompt()))
	b.WriteString("\n\n")
	if m.eng != nil {
		b.WriteString(m.renderBoard(false))
	}
	b.WriteString("\n")
	b.WriteString(statusSt.Render(m.status))
	return b.String()
}

func (m model) viewPlay() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("%s to play", m.eng.ToPlay())))
	b.WriteString("  ")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("prisoners B %d  W %d  komi %.1f",
		m.eng.Captured(engine.Black), m.eng.Captured(engine.White), m.eng.Komi())))
	b.WriteString("\n\n")
	b.WriteString(m.renderBoard(false))
	b.WriteString("\n")
	b.WriteString(statusSt.Render(m.status))
	return b.String()
}

func (m model) viewScore() string {
	prev := m.eng.Score(m.dead)
	var b strings.Builder
	b.WriteString(titleStyle.Render("Scoring  ·  mark dead stones"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("preview  Black %.1f   White %.1f", prev.Black, prev.White)))
	b.WriteString("\n\n")
	b.WriteString(m.renderBoard(true))
	b.WriteString("\n")
	b.WriteString(statusSt.Render(m.status))
	return b.String()
}

func (m model) viewResult() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.result.String()))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Japanese territory scoring"))
	b.WriteString("\n")
	if m.result.Resigned == engine.Empty {
		r := m.result
		b.WriteString(fmt.Sprintf(
			"\n            Black    White\n"+
				"Territory   %5d    %5d\n"+
				"Prisoners   %5d    %5d\n"+
				"Komi            —    %5.1f\n"+
				"Total       %5.1f    %5.1f\n",
			r.BlackTerritory, r.WhiteTerritory,
			r.BlackCaptures, r.WhiteCaptures,
			r.Komi,
			r.Black, r.White,
		))
	}
	b.WriteString("\n")
	if m.eng != nil {
		b.WriteString(m.renderBoard(true))
	}
	b.WriteString("\n")
	b.WriteString(statusSt.Render(m.status))
	return b.String()
}

func (m model) renderBoard(markDead bool) string {
	var b strings.Builder
	b.WriteString("   ")
	for x := 0; x < m.size; x++ {
		b.WriteString(fmt.Sprintf("%c ", 'A'+x))
	}
	b.WriteByte('\n')
	for y := 0; y < m.size; y++ {
		b.WriteString(fmt.Sprintf("%2d ", y+1))
		for x := 0; x < m.size; x++ {
			p := engine.Point{X: x, Y: y}
			glyph := m.cell(p, markDead)
			if p == m.cursor && (m.scene == scenePlay || m.scene == sceneScore) {
				b.WriteString(cursorSt.Render("[") + glyph + cursorSt.Render("]"))
				continue
			}
			b.WriteString(glyph)
			b.WriteByte(' ')
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func (m model) cell(p engine.Point, markDead bool) string {
	if markDead {
		if _, ok := m.dead[p]; ok {
			return deadSt.Render("×")
		}
	}
	switch m.eng.At(p) {
	case engine.Black:
		return blackSt.Render("●")
	case engine.White:
		return whiteSt.Render("○")
	default:
		for _, h := range engine.Hoshi(m.size) {
			if h == p {
				return "·"
			}
		}
		if lp := m.eng.LastPoint(); lp != nil && *lp == p {
			return "+"
		}
		return "."
	}
}
