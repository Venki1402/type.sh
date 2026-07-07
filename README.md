# type.sh

A terminal typing-speed trainer, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Live per-character highlighting, real-time WPM, persistent progress, and keystroke-level anti-cheat.

<img width="872" height="328" alt="Screenshot 2026-07-04 at 11 45 55 AM" src="https://github.com/user-attachments/assets/d20b1a78-3bfe-491b-863c-b295a38091db" />


## Features

- **Live typing** — every keystroke is captured and colored (correct / incorrect / pending) in real time, with a moving cursor and live WPM.
- **Time & word modes** — 15/30/45/60s or 10/25/50/100 words.
- **Global config** — your name, theme, and preferences are stored once (in your OS config dir) and reused on every run. No more re-entering setup.
- **Progress tracking** — XP, levels, best WPM/accuracy, and a recent-test history that survives restarts.
- **Keystroke-timing anti-cheat** — instead of only checking headline WPM, it inspects the *rhythm* of your typing (inter-key intervals, jitter, paste bursts) to flag automation and pasting.
- **Themes** — `default`, `warm`, `mono`, switchable in Settings.

## Install

**Homebrew** (macOS / Linux):

```
brew install Venki1402/tap/typesh
```

**Install script** (macOS / Linux, no Homebrew needed):

```
curl -fsSL https://raw.githubusercontent.com/Venki1402/type.sh/main/install.sh | sh
```

**Windows**:

Download the latest `typesh_*_windows_amd64.zip` (or `_arm64`) from the
[Releases](https://github.com/Venki1402/type.sh/releases) page, unzip it, and run
`typesh.exe` from Windows Terminal or PowerShell.

Then just run:

```
typesh
```

None of these methods require Go — they install a prebuilt binary.

### Build from source

If you'd rather build it yourself (requires Go 1.24+):

```
git clone https://github.com/Venki1402/type.sh
cd type.sh
go run .          # or: go build -o typesh . && ./typesh
```

### Controls

| Screen   | Keys                                   |
|----------|----------------------------------------|
| Menu     | `↑ ↓` move · `enter` select · `q` quit |
| Mode     | `← →` choose · `enter` start · `esc` back |
| Typing   | just type · `backspace` fix · `esc` cancel |
| Result   | `r` retry · any key for menu           |
| Settings | `↑ ↓` move · `← → / enter` change · `esc` save |

## Config location

Settings live in your OS config directory under `type.sh/config.json`
(e.g. `~/Library/Application Support/type.sh/` on macOS,
`~/.config/type.sh/` on Linux, `%AppData%\type.sh\` on Windows).
The Stats screen shows the exact path.

## Project layout

```
main.go                 entry point
internal/config         global persisted config, profile & history
internal/words          word bank + passage generation
internal/typing         keystroke capture + WPM/accuracy/consistency
internal/cheat          keystroke-timing anti-cheat
internal/tui            Bubble Tea screens + lipgloss theming
```

## Want to contribute?

- Open an issue describing the bug fix / feature.
- We'll discuss how to solve it.
- Then send a PR — I'll review and merge it. Good luck!
