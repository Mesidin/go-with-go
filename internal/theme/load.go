package theme

import (
	"bufio"
	"image/color"
	"os"
	"path/filepath"
	"strings"
)

func candidateDirs() []string {
	if d := os.Getenv("OMARCHY_THEME_DIR"); d != "" {
		return []string{d}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	dirs := []string{
		filepath.Join(home, ".local", "state", "omarchy", "current", "theme"),
		filepath.Join(home, ".config", "omarchy", "current", "theme"),
	}
	namePath := filepath.Join(home, ".local", "state", "omarchy", "current", "theme.name")
	if b, err := os.ReadFile(namePath); err == nil {
		name := strings.TrimSpace(string(b))
		if name != "" {
			dirs = append(dirs,
				filepath.Join(home, ".config", "omarchy", "themes", name),
				filepath.Join("/usr", "share", "omarchy", "themes", name),
			)
		}
	}
	return dirs
}

func loadFromDir(dir string) (Palette, bool) {
	if p, ok := parseColorsTOML(filepath.Join(dir, "colors.toml")); ok {
		p.Name = filepath.Base(dir)
		return finish(p), true
	}
	if p, ok := parseAlacritty(filepath.Join(dir, "alacritty.toml")); ok {
		p.Name = filepath.Base(dir)
		return finish(p), true
	}
	return Palette{}, false
}

func parseColorsTOML(path string) (Palette, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Palette{}, false
	}
	defer f.Close()
	kv := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		k, v, ok := splitKV(line)
		if !ok {
			continue
		}
		kv[k] = v
	}
	if len(kv) == 0 {
		return Palette{}, false
	}
	return paletteFromMap(kv), true
}

func parseAlacritty(path string) (Palette, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Palette{}, false
	}
	defer f.Close()
	section := ""
	kv := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			continue
		}
		k, v, ok := splitKV(line)
		if !ok {
			continue
		}
		switch section {
		case "colors.primary":
			if k == "background" || k == "foreground" {
				kv[k] = v
			}
			if k == "dim_foreground" {
				kv["muted"] = v
			}
		case "colors.cursor":
			if k == "cursor" {
				kv["accent"] = v
			}
		case "colors.selection":
			if k == "background" {
				kv["selection"] = v
			}
		case "colors.normal":
			kv["color"+ansiIndex(k)] = v
			if k == "red" {
				kv["red"] = v
			}
		case "colors.bright":
			if k == "white" {
				kv["color15"] = v
			}
			if k == "black" {
				kv["color8"] = v
			}
		}
	}
	if kv["background"] == "" && kv["foreground"] == "" {
		return Palette{}, false
	}
	return paletteFromMap(kv), true
}

func ansiIndex(name string) string {
	switch name {
	case "black":
		return "0"
	case "red":
		return "1"
	case "green":
		return "2"
	case "yellow":
		return "3"
	case "blue":
		return "4"
	case "magenta":
		return "5"
	case "cyan":
		return "6"
	case "white":
		return "7"
	default:
		return ""
	}
}

func splitKV(line string) (string, string, bool) {
	i := strings.IndexByte(line, '=')
	if i < 0 {
		return "", "", false
	}
	k := strings.TrimSpace(line[:i])
	v := strings.TrimSpace(line[i+1:])
	v = strings.Trim(v, `"'`)
	if k == "" || v == "" {
		return "", "", false
	}
	return k, v, true
}

func paletteFromMap(kv map[string]string) Palette {
	hex := func(keys ...string) color.RGBA {
		for _, k := range keys {
			if v, ok := kv[k]; ok {
				if c, ok := parseHex(v); ok {
					return c
				}
			}
		}
		return color.RGBA{}
	}
	return Palette{
		Background:        hex("background", "bg"),
		DarkBackground:    hex("dark_background", "dark_bg"),
		LighterBackground: hex("lighter_background", "lighter_bg"),
		Foreground:        hex("foreground", "fg"),
		DarkForeground:    hex("dark_foreground", "dark_fg"),
		Accent:            hex("accent", "cursor", "color4"),
		Muted:             hex("muted", "color8", "color6"),
		Selection:         hex("selection", "selection_background"),
		Red:               hex("red", "color1"),
		Color7:            hex("color7", "bright_foreground"),
	}
}
