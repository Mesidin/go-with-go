package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"go-with-go/engine"
	"go-with-go/internal/theme"
)

var (
	colWood      = color.RGBA{0xDC, 0xB3, 0x5C, 0xFF}
	colWoodDark  = color.RGBA{0xC4, 0x98, 0x48, 0xFF}
	colGrid      = color.RGBA{0x3A, 0x24, 0x12, 0xFF}
	colBlack     = color.RGBA{0x14, 0x12, 0x10, 0xFF}
	colBlackHi   = color.RGBA{0x4A, 0x4A, 0x4A, 0xFF}
	colWhite     = color.RGBA{0xF4, 0xF0, 0xE6, 0xFF}
	colWhiteEdge = color.RGBA{0xC8, 0xC0, 0xB0, 0xFF}
	colLast      = color.RGBA{0xC8, 0x32, 0x28, 0xFF}
	colHoverB    = color.RGBA{0x14, 0x12, 0x10, 0x66}
	colHoverW    = color.RGBA{0xF4, 0xF0, 0xE6, 0x66}
	colKo        = color.RGBA{0x8B, 0x1A, 0x1A, 0xFF}
	colDeadMark  = color.RGBA{0xE8, 0x5D, 0x04, 0xFF}
	colFlash     = color.RGBA{0xE0, 0x30, 0x30, 0xAA}
	colHUD       color.RGBA
	colPanel     color.RGBA
	colBtn       color.RGBA
	colBtnHot    color.RGBA
	colBtnBorder color.RGBA
	colBtnText   color.RGBA
	colMuted     color.RGBA
	colInk       color.RGBA
	colWin       color.RGBA
	colTerrB     = color.RGBA{0x1A, 0x18, 0x16, 0xB8}
	colTerrW     = color.RGBA{0xFF, 0xF6, 0xE8, 0xC8}
	colDame      = color.RGBA{0x8A, 0x70, 0x48, 0x99}
)

func init() {
	applyChrome(theme.Current())
}

func applyChrome(p theme.Palette) {
	colHUD = p.Background
	colPanel = p.LighterBackground
	colBtn = theme.Mix(p.Background, p.Foreground, 0.32)
	colBtnHot = p.Accent
	colBtnBorder = p.Muted
	if colBtnBorder.R == colHUD.R && colBtnBorder.G == colHUD.G && colBtnBorder.B == colHUD.B {
		colBtnBorder = p.Accent
	}
	colBtnText = p.Foreground
	colMuted = p.Muted
	colInk = p.Foreground
	colWin = p.Accent
	colLast = p.Red
	colKo = p.Red
	colFlash = p.Red
	colFlash.A = 0xAA
	colDeadMark = p.Accent
}

type boardView struct {
	originX, originY float32
	cell             float32
	size             int
}

func layoutBoard(screenW, screenH, hudH, size int) boardView {
	return layoutBoardPad(screenW, screenH, 24, hudH, size)
}

func layoutBoardPad(screenW, screenH, top, bottom, size int) boardView {
	availW := float32(screenW) - 48
	availH := float32(screenH-bottom-top) - 24
	if availH < 80 {
		availH = 80
	}
	cell := availW / float32(size)
	if availH/float32(size) < cell {
		cell = availH / float32(size)
	}
	board := cell * float32(size)
	return boardView{
		originX: (float32(screenW) - board) / 2,
		originY: float32(top),
		cell:    cell,
		size:    size,
	}
}

func (b boardView) pointCenter(p engine.Point) (float32, float32) {
	return b.originX + (float32(p.X)+0.5)*b.cell,
		b.originY + (float32(p.Y)+0.5)*b.cell
}

func (b boardView) hit(x, y int) (engine.Point, bool) {
	fx := (float32(x) - b.originX) / b.cell
	fy := (float32(y) - b.originY) / b.cell
	if fx < -0.3 || fy < -0.3 || fx > float32(b.size)+0.3 || fy > float32(b.size)+0.3 {
		return engine.Point{}, false
	}
	px := int(math.Floor(float64(fx)))
	py := int(math.Floor(float64(fy)))
	if px < 0 || py < 0 || px >= b.size || py >= b.size {
		return engine.Point{}, false
	}
	return engine.Point{X: px, Y: py}, true
}

func (b boardView) draw(dst *ebiten.Image, g *engine.Game, hover *engine.Point, dead map[engine.Point]struct{}, flash *engine.Point, marks []engine.Point, terr map[engine.Point]engine.Color) {
	board := b.cell * float32(b.size)
	vector.FillRect(dst, b.originX-8, b.originY-8, board+16, board+16, colWoodDark, true)
	vector.FillRect(dst, b.originX, b.originY, board, board, colWood, true)

	// Grid is drawn on the intersections, inset by half a cell.
	inset := b.cell * 0.5
	n := float32(b.size - 1)
	for i := 0; i < b.size; i++ {
		off := inset + float32(i)*b.cell
		vector.StrokeLine(dst, b.originX+inset, b.originY+off, b.originX+inset+n*b.cell, b.originY+off, 1.4, colGrid, true)
		vector.StrokeLine(dst, b.originX+off, b.originY+inset, b.originX+off, b.originY+inset+n*b.cell, 1.4, colGrid, true)
	}

	rHoshi := b.cell * 0.08
	for _, h := range engine.Hoshi(b.size) {
		cx, cy := b.pointCenter(h)
		vector.FillCircle(dst, cx, cy, rHoshi, colGrid, true)
	}

	if ko := g.Ko(); ko != nil {
		cx, cy := b.pointCenter(*ko)
		vector.StrokeCircle(dst, cx, cy, b.cell*0.18, 2, colKo, true)
	}

	if terr != nil {
		tr := b.cell * 0.20
		for y := 0; y < b.size; y++ {
			for x := 0; x < b.size; x++ {
				p := engine.Point{X: x, Y: y}
				cx, cy := b.pointCenter(p)
				if own, ok := terr[p]; ok {
					col := colTerrB
					if own == engine.White {
						col = colTerrW
					}
					vector.FillCircle(dst, cx, cy, tr, col, true)
					continue
				}
				if g.At(p) == engine.Empty {
					if _, marked := dead[p]; !marked {
						vector.FillCircle(dst, cx, cy, tr*0.45, colDame, true)
					}
				}
			}
		}
	}

	r := b.cell * 0.44
	for y := 0; y < b.size; y++ {
		for x := 0; x < b.size; x++ {
			p := engine.Point{X: x, Y: y}
			c := g.At(p)
			if c == engine.Empty {
				continue
			}
			cx, cy := b.pointCenter(p)
			drawStone(dst, cx, cy, r, c)
			if _, ok := dead[p]; ok {
				vector.StrokeLine(dst, cx-r*0.5, cy-r*0.5, cx+r*0.5, cy+r*0.5, 3, colDeadMark, true)
				vector.StrokeLine(dst, cx+r*0.5, cy-r*0.5, cx-r*0.5, cy+r*0.5, 3, colDeadMark, true)
			}
		}
	}

	if lp := g.LastPoint(); lp != nil && g.At(*lp) != engine.Empty {
		cx, cy := b.pointCenter(*lp)
		vector.FillCircle(dst, cx, cy, r*0.18, colLast, true)
	}

	for _, p := range marks {
		cx, cy := b.pointCenter(p)
		vector.StrokeCircle(dst, cx, cy, b.cell*0.28, 3, colWin, true)
	}

	if hover != nil && g.At(*hover) == engine.Empty {
		cx, cy := b.pointCenter(*hover)
		col := colHoverB
		if g.ToPlay() == engine.White {
			col = colHoverW
		}
		vector.FillCircle(dst, cx, cy, r, col, true)
	}

	if flash != nil {
		cx, cy := b.pointCenter(*flash)
		vector.FillCircle(dst, cx, cy, r, colFlash, true)
	}
}

func drawStone(dst *ebiten.Image, cx, cy, r float32, c engine.Color) {
	if c == engine.Black {
		vector.FillCircle(dst, cx+1.2, cy+1.6, r, color.RGBA{0, 0, 0, 50}, true)
		vector.FillCircle(dst, cx, cy, r, colBlack, true)
		vector.FillCircle(dst, cx-r*0.28, cy-r*0.28, r*0.28, colBlackHi, true)
		return
	}
	vector.FillCircle(dst, cx+1.2, cy+1.6, r, color.RGBA{0, 0, 0, 40}, true)
	vector.FillCircle(dst, cx, cy, r, colWhite, true)
	vector.StrokeCircle(dst, cx, cy, r, 1.3, colWhiteEdge, true)
}
