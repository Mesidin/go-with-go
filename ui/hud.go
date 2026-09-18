package ui

import (
	"bytes"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	face    *text.GoTextFace
	faceBig *text.GoTextFace
	faceSm  *text.GoTextFace
)

func init() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	face = &text.GoTextFace{Source: src, Size: 18}
	faceBig = &text.GoTextFace{Source: src, Size: 32}
	faceSm = &text.GoTextFace{Source: src, Size: 14}
}

type button struct {
	x, y, w, h float32
	label      string
	id         string
}

func (b button) contains(x, y int) bool {
	fx, fy := float32(x), float32(y)
	return fx >= b.x && fy >= b.y && fx < b.x+b.w && fy < b.y+b.h
}

func drawButton(dst *ebiten.Image, b button, mx, my int) {
	fill, border := colBtn, colBtnBorder
	if b.contains(mx, my) {
		fill = colBtnHot
		border = colBtnText
	}
	vector.FillRect(dst, b.x, b.y, b.w, b.h, fill, true)
	vectorStrokeRect(dst, b.x, b.y, b.w, b.h, border)
	drawLabel(dst, b.label, b.x+b.w/2, b.y+b.h/2-9, face, colBtnText, true)
}

func drawLabel(dst *ebiten.Image, s string, x, y float32, f *text.GoTextFace, col color.Color, center bool) {
	align := text.AlignStart
	if center {
		align = text.AlignCenter
	}
	drawLabelAlign(dst, s, x, y, f, col, align)
}

func drawLabelAlign(dst *ebiten.Image, s string, x, y float32, f *text.GoTextFace, col color.Color, align text.Align) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(col)
	op.PrimaryAlign = align
	text.Draw(dst, s, f, op)
}

func drawLabelRight(dst *ebiten.Image, s string, x, y float32, f *text.GoTextFace, col color.Color) {
	drawLabelAlign(dst, s, x, y, f, col, text.AlignEnd)
}

func drawPanel(dst *ebiten.Image, x, y, w, h float32) {
	vector.FillRect(dst, x, y, w, h, colHUD, true)
}

func wrapWords(s string, maxPx float32, f *text.GoTextFace) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		try := cur + " " + w
		width, _ := text.Measure(try, f, 0)
		if float32(width) > maxPx {
			lines = append(lines, cur)
			cur = w
			continue
		}
		cur = try
	}
	return append(lines, cur)
}

func drawWrapped(dst *ebiten.Image, s string, x, y, maxW, lineH float32, f *text.GoTextFace, col color.Color) {
	for i, line := range wrapWords(s, maxW, f) {
		drawLabel(dst, line, x, y+float32(i)*lineH, f, col, false)
	}
}
