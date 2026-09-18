# Go with Go

A game of Go written in Go. Japanese territory scoring, simple ko, 9×9 / 13×13 / 19×19, two humans on one machine or a capture-aware bot, plus a learn-to-play course.

One rules engine, three ways to play: a native window, the same UI in a browser via WebAssembly, and a terminal UI. There is no game server. Browser play runs the engine in the client.

© 2026 Bradley Erickson

## Requirements

- Go 1.26 or later ([go.dev/dl](https://go.dev/dl/)). Ebitengine v2.10 is pure Go on desktop, so you do not need a C compiler.

Menus, HUD, and the TUI follow the **active Omarchy theme** (`colors.toml` under `~/.local/state/omarchy/current/theme`). Off Omarchy, or before you pick a theme, the chrome is [Gold Rush](https://github.com/tahayvr/omarchy-gold-rush-theme). Switching themes with `omarchy theme set` retints a running window. The **goban stays wood** on every OS.
- A graphics driver for the windowed app (any ordinary desktop).
- On Omarchy / other Wayland Linux: XWayland (Omarchy includes it).

## Install and build

Clone the repo, then follow the section for your OS. In every case, run commands from the repo root.

```
git clone <this-repo-url> go-with-go
cd go-with-go
```

### Windows

1. Install Go if needed: `winget install --id GoLang.Go --exact`, or the MSI from [go.dev/dl](https://go.dev/dl/).
2. Open a **new** PowerShell window so `go` is on PATH.
3. Check the install: `go version` (needs 1.26 or later).
4. Run tests: `go test ./...`
5. Play:

```
go run .\cmd\go-with-go
```

PowerShell does not run a program in the current folder by name. From the repo you can also use:

```
go build -o go-with-go.exe .\cmd\go-with-go
.\go-with-go.exe
```

or `.\play.ps1`.

To install onto PATH (`%USERPROFILE%\go\bin`):

```
go install .\cmd\go-with-go
go install .\cmd\tui
```

If `go-with-go` is still not found, add Go and the Go bin directory to this session, then open a new terminal later:

```
$env:Path += ";C:\Program Files\Go\bin;$env:USERPROFILE\go\bin"
```

If the window opens and closes, read `go-with-go.log` next to the executable.

### macOS

1. Install Go with Homebrew (`brew install go`) or the package from [go.dev/dl](https://go.dev/dl/).
2. Check the install: `go version` (needs 1.26 or later).
3. Run tests: `go test ./...`
4. Play:

```
go run ./cmd/go-with-go
```

Build a binary:

```
go build -o go-with-go ./cmd/go-with-go
./go-with-go
```

To put it in the **Dock**, build an app bundle (on the Mac, from the repo root):

```
sh scripts/macos-app.sh
cp -R "Go with Go.app" /Applications/
open /Applications
```

Drag **Go with Go** onto the Dock. The first time macOS may ask you to right-click the app and choose Open. A raw `go-with-go` binary can also be dragged to the Dock, but a `.app` in `/Applications` is what the Dock is meant for.

Install onto PATH (`$(go env GOPATH)/bin`, usually `~/go/bin`):

```
go install ./cmd/go-with-go
go install ./cmd/tui
```

If the command is not found, add that bin directory to PATH in `~/.zprofile`:

```
export PATH="$(go env GOPATH)/bin:$PATH"
```

### Omarchy (Arch Linux)

Omarchy is Arch-based and uses Hyprland. The windowed game talks to X11, so it runs under XWayland.

1. Install Go and the X11/OpenGL libraries (Mesa is typical):

```
sudo pacman -S go libx11 libglvnd mesa
```

Optional, recommended for cursor and multi-monitor: `libxcursor libxi libxinerama libxrandr libxrender libxext`.

2. Check the install: `go version` (needs 1.26 or later).
3. Run tests: `go test ./...`
4. Play:

```
go run ./cmd/go-with-go
```

If the window does not open, use the terminal UI:

```
go run ./cmd/tui
```

Build binaries:

```
go build -o go-with-go ./cmd/go-with-go
go build -o go-tui ./cmd/tui
```

Install onto PATH (`$(go env GOPATH)/bin`):

```
go install ./cmd/go-with-go
go install ./cmd/tui
```

Add `export PATH="$(go env GOPATH)/bin:$PATH"` to your shell config if the commands are not found.

## How to play

The home screen is Start game, Learn to play, Tsumego, Load SGF, and Quit. Start game opens setup (board, humans vs bot, handicap, komi, clock). Bot easy/club appears only if you pick Black vs bot or White vs bot.

Click an intersection to place a stone. **P** pass, **U** undo, **R** resign, **Esc** menu. After two passes, click groups to mark them dead. Empty points shade as territory: dark for Black, light for White, small brown for dame. Stones on the board do not count.

**Learn to play** (menu button, or `t` in the TUI) walks through placing, liberties, capture, suicide, ko, two eyes, and Japanese scoring on the real board. Skip and Back move between lessons.

**Tsumego** is a set of life-and-death exercises. Play the vital point; a wrong legal move is taken back so you can try again.

**SGF** (Smart Game Format) is the usual `.sgf` text file for a Go game: size, komi, handicap, and every move. **Save (S)** writes `games/YYYYMMDD-HHMMSS.sgf`. **Load SGF** lists that folder. You can also drop an `.sgf` file on the window.

**Moves (M)** shows the move list with coordinates (A–T, skipping I) and capture comments.

Handicap 2–9 places Black on the star points and White plays first (komi switches to 0.5 unless you change it). The clock is simple sudden death per player. **Bot club** adds self-atari and one-reply capture checks; **easy** is the original sparring partner.

Vs the bot, undo takes back the bot reply and your move together.

### Terminal

```
go run ./cmd/tui
```

Enter opens setup, then Enter starts. `t` learn, `g` tsumego, `q` quit. In play: arrows or `hjkl`, Enter to play, `p` pass, `u` undo, `r` resign, `s` save SGF. In scoring, Enter marks dead, `c` confirms, `x` resumes.

### Browser (no game server)

Build WebAssembly, copy the Go wasm loader, then serve the `web/` folder. Browsers will not load WASM from `file://`.

Windows PowerShell:

```
$env:GOOS = 'js'; $env:GOARCH = 'wasm'
go build -o web/game.wasm ./cmd/go-with-go
Remove-Item Env:GOOS, Env:GOARCH
Copy-Item "$(go env GOROOT)\lib\wasm\wasm_exec.js" web\
go run ./cmd/serve
```

macOS and Omarchy:

```
GOOS=js GOARCH=wasm go build -o web/game.wasm ./cmd/go-with-go
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/
go run ./cmd/serve
```

Open http://127.0.0.1:8080. `cmd/serve` only serves static files; the game still runs in the browser.

## Rules

- Place, capture by liberty, suicide illegal (capturing “suicide” allowed)
- Simple ko
- Two consecutive passes end the game; resign is allowed
- **Japanese scoring:** empty points surrounded by one color, plus prisoners (captures and groups marked dead at the end), plus **6.5 komi** to White on every board size
- Dame (empty points touching both colors) score for neither
- Stones remaining on the board do not count
- Auto-score is a flood fill, not a life-and-death reader. Mark dead groups, or play the position out, if seki looks wrong

The bot is a 9×9 sparring partner, not a ranked engine.

## Layout

```
engine/   rules, Japanese scoring, SGF, handicap (no UI)
bot/      easy and club heuristics
learn/    interactive beginner course
tsumego/  life-and-death problems
ui/       Ebitengine window (desktop + WASM)
tui/      Bubble Tea terminal UI
cmd/go-with-go
cmd/tui
cmd/serve
web/      index.html + WASM build output
```
