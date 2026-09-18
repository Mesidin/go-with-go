# Go with Go — agent notes

A playable game of Go written in Go. Japanese territory scoring, Ebitengine window, Bubble Tea TUI, learn-to-play, tsumego, SGF.

## Layout

- `engine/` — rules, scoring, SGF, handicap. No UI imports.
- `bot/` — easy / club heuristics.
- `learn/`, `tsumego/` — courses and problems.
- `ui/` — Ebitengine desktop + WASM.
- `tui/` — terminal UI.
- `internal/theme/` — Omarchy palette loader.
- `internal/meta/` — copyright and title.

## Run

```
go test ./...
go run ./cmd/go-with-go
go run ./cmd/tui
```

On Windows PowerShell use `.\cmd\...`. Do not kill a running game window to rebuild unless the user asks.

macOS Dock: `sh scripts/macos-app.sh` then copy `Go with Go.app` to `/Applications` and drag it to the Dock. Build that script on a Mac (or cross-compile `GOOS=darwin`).

## Omarchy theming

Menus, HUD, buttons, and TUI colors follow the **active Omarchy theme**.

1. Read `colors.toml` (else `alacritty.toml`) from, in order:
   - `$OMARCHY_THEME_DIR`
   - `~/.local/state/omarchy/current/theme`
   - `~/.config/omarchy/current/theme`
   - `~/.config/omarchy/themes/<theme.name>`
2. If none of those exist, use the baked-in **Gold Rush** palette (`internal/theme.GoldRush`, [tahayvr/omarchy-gold-rush-theme](https://github.com/tahayvr/omarchy-gold-rush-theme)).
3. Reload about once a second so `omarchy theme set` retints a running game.

**Do not theme the goban.** Wood (`colWood`, grid, stones) stays wooden on every OS. Only chrome (window fill, panels, buttons, labels, selection outline) uses the palette.

Map: background → HUD, lighter_background → buttons, accent → selected/outline/headings, foreground → body text, muted → secondary text, red → last-move / ko / flash.

## Product rules

- Japanese scoring: land + prisoners + 6.5 komi (unless the user set komi). Stones on the board do not count.
- Home screen is Start / Learn / Tsumego / Load SGF / Quit. Setup (board, humans vs bot, handicap, clock) is a second screen. Bot easy/club only when vs bot.
- Copyright: © 2026 Bradley Erickson (`internal/meta`).
- Do not invent network play or a ranked engine.

## Tests

`go test ./...` must stay green. Engine tests cover capture, ko, Japanese score, SGF.
