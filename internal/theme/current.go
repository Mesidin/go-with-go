package theme

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	cached  Palette
	stamp   string
	checked time.Time
)

func init() {
	cached = resolve(GoldRush())
}

// Current returns the active Omarchy palette, or Gold Rush if none is installed.
func Current() Palette {
	mu.Lock()
	defer mu.Unlock()
	return cached
}

// Tick reloads the Omarchy theme if it changed. Call from the UI loop.
func Tick() Palette {
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	if now.Sub(checked) < time.Second {
		return cached
	}
	checked = now
	s := fingerprint()
	if s == stamp && stamp != "" {
		return cached
	}
	stamp = s
	cached = resolve(GoldRush())
	return cached
}

func resolve(fallback Palette) Palette {
	for _, dir := range candidateDirs() {
		if p, ok := loadFromDir(dir); ok {
			return p
		}
	}
	return fallback
}

func fingerprint() string {
	var b strings.Builder
	for _, dir := range candidateDirs() {
		for _, name := range []string{"colors.toml", "alacritty.toml"} {
			path := filepath.Join(dir, name)
			if st, err := os.Stat(path); err == nil {
				b.WriteString(path)
				b.WriteByte('|')
				b.WriteString(strconv.FormatInt(st.ModTime().UnixNano(), 10))
				b.WriteByte(';')
			}
		}
	}
	home, err := os.UserHomeDir()
	if err == nil {
		p := filepath.Join(home, ".local", "state", "omarchy", "current", "theme.name")
		if st, err := os.Stat(p); err == nil {
			b.WriteString(p)
			b.WriteByte('|')
			b.WriteString(strconv.FormatInt(st.ModTime().UnixNano(), 10))
		}
	}
	return b.String()
}
