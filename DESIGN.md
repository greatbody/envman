# envman — Environment Variable Profile Manager

A CLI tool for managing named environment variable profiles. Load different sets of environment variables into your shell session on demand.

## User Stories

| ID | Story |
|----|-------|
| US1 | As a user, I can set a default profile so that every new shell start automatically injects those env vars |
| US2 | As a user, I can dynamically load a profile into my current terminal session without changing the default profile |
| US3 | As a user, I can launch a TUI to manage all profiles — switch default, load dynamically, create, edit, delete, view, diff, copy |

## Storage

```
~/.envman/
├── profiles/
│   ├── default.env      # created on first run
│   ├── work.env
│   └── project-xyz.env
└── config.json          # { "default_profile": "default" }
```

Each profile is a standard `.env` file, compatible with docker-compose, VSCode, direnv, etc.

## Commands

| Command | Output to shell? | Description |
|---------|-----------------|-------------|
| `envman init` | Yes | Shell init snippet — loads default profile |
| `envman list` | No | List all profiles, mark default with `*` |
| `envman create <name>` | No | Create new profile (opens `$EDITOR`) |
| `envman show <name>` | No | Print profile contents |
| `envman load <name>` | Yes (`export`) | Inject profile vars into current shell |
| `envman unload [name]` | Yes (`unset`) | Unload last-loaded, specific, or `--all` |
| `envman tui` | Yes (if user loads) | Open TUI for full management |
| `envman default <name>` | No | Set default profile (takes effect next shell) |
| `envman diff <a> <b>` | No | Compare two profiles side by side |
| `envman copy <src> <dst>` | No | Duplicate a profile |
| `envman delete <name>` | No | Delete a profile |

## Stackable Loads

Multiple profiles can be loaded in sequence. Later vars override earlier ones on key conflict.

```bash
eval $(envman load base)     # loads base vars
eval $(envman load work)     # adds work vars, overrides conflicts
eval $(envman unload work)   # unloads only work, base remains
eval $(envman unload --all)  # clears all dynamically loaded
```

**Tracking**: An env var `ENVMAN_LOADED_PROFILES` (comma-separated, in load order) tracks what's loaded in the current shell session.

## Shell Setup (.zshrc)

```bash
# envman default profile
eval $(envman init)

# shell function for seamless load/unload/tui
envman() {
  case "$1" in
    load|unload|init|tui)
      eval "$(command envman "$@")"
      ;;
    *)
      command envman "$@"
      ;;
  esac
}
```

## TUI Layout (bubbletea)

```
┌─ envman ──────────────────────────────────────────┐
│                                                    │
│  ★ default        ← currently default             │
│    work            ← loaded dynamically [loaded]   │
│    project-xyz                                  │
│    personal                                      │
│                                                    │
│────────────────────────────────────────────────────│
│  [s] set default  [l] load  [e] edit  [n] new     │
│  [d] delete       [v] view  [c] copy  [q] quit    │
│  [D] diff                                       │
└────────────────────────────────────────────────────┘
```

- Arrow keys navigate, shortcut keys trigger actions
- `s` switches default (persists to config.json)
- `l` loads profile into session (outputs export on TUI exit)
- `e` opens `$EDITOR` for selected profile
- `v` shows profile contents in a detail pane
- `d` deletes with confirmation
- `D` enters diff mode (select two profiles)
- Multiple profiles can be tagged `[loaded]` for stackable loads

## Project Structure

```
envman/
├── cmd/
│   └── envman/
│       └── main.go
├── internal/
│   ├── cli/           # cobra command definitions
│   ├── profile/       # profile CRUD, parsing .env files
│   ├── shell/         # export/unset output generation
│   ├── config/        # config.json management
│   └── tui/           # bubbletea TUI
│       ├── model.go
│       ├── update.go
│       ├── view.go
│       └── keys.go
├── go.mod
├── Makefile
└── README.md
```

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — TUI styling
- `github.com/charmbracelet/bubbles` — TUI components (list, text input)

## Build & Install

```bash
make build          # builds to ./bin/envman
make install        # copies to /usr/local/bin/envman
```

## Key Behaviors

1. **Default profile** is loaded transparently on shell start via `eval $(envman init)` in `.zshrc`
2. **Dynamic loads** are ephemeral — they don't modify any config file, just output `export` statements
3. **Stackable** — multiple loads layer on each other, tracked via `ENVMAN_LOADED_PROFILES` env var
4. **TUI** does everything — management AND loading. On exit, outputs any load/unload commands the user triggered
5. **Plain .env files** — compatible with docker-compose, VSCode, direnv, etc.
