package theme

import (
	"fmt"
	"image/color"
)

// Palette is the Omarchy colors.toml set used for chrome (and chess boards).
type Palette struct {
	Name string

	Background        color.RGBA
	DarkBackground    color.RGBA
	LighterBackground color.RGBA
	Foreground        color.RGBA
	DarkForeground    color.RGBA
	Accent            color.RGBA
	Muted             color.RGBA
	Selection         color.RGBA
	Red               color.RGBA
	Color7            color.RGBA
}

func (p Palette) Hex(c color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

func (p Palette) Lipgloss(c color.RGBA) string {
	return p.Hex(c)
}

// Mix blends a toward b by t in 0..1.
func Mix(a, b color.RGBA, t float64) color.RGBA { return mix(a, b, t) }

func mix(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 255,
	}
}

func withAlpha(c color.RGBA, a uint8) color.RGBA {
	c.A = a
	return c
}

func mustHex(s string) color.RGBA {
	c, ok := parseHex(s)
	if !ok {
		return color.RGBA{0, 0, 0, 255}
	}
	return c
}

func parseHex(s string) (color.RGBA, bool) {
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	var n uint32
	switch len(s) {
	case 6:
		_, err := fmt.Sscanf(s, "%06x", &n)
		if err != nil {
			return color.RGBA{}, false
		}
		return color.RGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 255}, true
	case 3:
		_, err := fmt.Sscanf(s, "%03x", &n)
		if err != nil {
			return color.RGBA{}, false
		}
		return color.RGBA{
			uint8((n>>8)&0xf) * 17,
			uint8((n>>4)&0xf) * 17,
			uint8(n&0xf) * 17,
			255,
		}, true
	default:
		return color.RGBA{}, false
	}
}

func finish(p Palette) Palette {
	if p.Foreground.A == 0 {
		p.Foreground = mustHex("#d9d9d9")
	}
	if p.Background.A == 0 {
		p.Background = mustHex("#121212")
	}
	if p.Accent.A == 0 {
		p.Accent = p.Foreground
	}
	if p.Muted.A == 0 {
		p.Muted = mix(p.Background, p.Foreground, 0.45)
	}
	if p.DarkBackground.A == 0 {
		p.DarkBackground = mix(p.Background, color.RGBA{}, 0.25)
	}
	if p.LighterBackground.A == 0 {
		p.LighterBackground = mix(p.Background, p.Foreground, 0.14)
	}
	if p.DarkForeground.A == 0 {
		p.DarkForeground = mix(p.Background, p.Foreground, 0.55)
	}
	if p.Selection.A == 0 {
		p.Selection = p.Accent
	}
	if p.Red.A == 0 {
		p.Red = p.Accent
	}
	if p.Color7.A == 0 {
		p.Color7 = p.Foreground
	}
	return p
}

// GoldRush is tahayvr's Omarchy Gold Rush theme (default for Go with Go).
// https://github.com/tahayvr/omarchy-gold-rush-theme
func GoldRush() Palette {
	return finish(Palette{
		Name:               "gold-rush",
		Background:         mustHex("#121212"),
		Foreground:         mustHex("#D9D9D9"),
		Accent:             mustHex("#C9A227"),
		Muted:              mustHex("#805B10"),
		Selection:          mustHex("#926C15"),
		Red:                mustHex("#DBB42C"),
		Color7:             mustHex("#EDC531"),
		DarkBackground:     mustHex("#0a0a0a"),
		LighterBackground:  mustHex("#1c1c1c"),
		DarkForeground:     mustHex("#805B10"),
	})
}

// HarborDark is HANCORE's Omarchy Harbor Dark theme (default for Chess with Go).
// https://github.com/HANCORE-linux/omarchy-harbordark-theme
func HarborDark() Palette {
	return finish(Palette{
		Name:               "harbordark",
		Background:         mustHex("#1B1B1B"),
		Foreground:         mustHex("#efebdc"),
		Accent:             mustHex("#e75a50"),
		Muted:              mustHex("#a99b7a"),
		Selection:          mustHex("#e75a50"),
		Red:                mustHex("#F44336"),
		Color7:             mustHex("#E1CE98"),
		DarkBackground:     mustHex("#141414"),
		LighterBackground:  mustHex("#2a2a2a"),
		DarkForeground:     mustHex("#817f68"),
	})
}
