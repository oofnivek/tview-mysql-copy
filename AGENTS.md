# AGENTS.md

## What this is

A Go terminal UI application built with [tview](https://github.com/rivo/tview) for managing and connecting to MySQL databases. Connections and preferences are stored locally in `~/.mysql-copy/`.

---

## Project layout

```
tview-mysql-copy/
├── main.go              Entry point — loads config/prefs, applies theme, starts tview app
├── config/
│   ├── config.go        Connection list: load/save ~/.mysql-copy/config.json
│   └── prefs.go         User preferences: load/save ~/.mysql-copy/prefs.json (theme)
└── ui/
    ├── connections.go   Connection list screen — navigation, theme toggle, CRUD actions
    ├── form.go          Add/edit connection form — field navigation, save/cancel
    └── theme.go         Dark and light tview.Theme definitions + ApplyTheme()
```

---

## Data files (`~/.mysql-copy/`)

| File | Purpose |
|------|---------|
| `config.json` | Saved MySQL connections (name, host, port, user, password, database) |
| `prefs.json` | User preferences — currently only `theme` (`"dark"` or `"light"`) |

Both files are created automatically on first run with safe permissions (`0600` files, `0700` directory).

---

## Key design decisions

- **`tview.Styles` is global** — `ApplyTheme()` in `ui/theme.go` overwrites it directly. After changing it, `app.Sync()` (not `app.ForceDraw()`) is required to force tcell to repaint all terminal cells, since tcell skips cells it considers unchanged when only colors differ.
- **Connections screen** (`ui/connections.go`) owns the `*tview.Pages` so the form can be pushed/popped on top of the list without recreating it.
- **Form navigation** (`ui/form.go`) intercepts `↑`/`↓` and converts them to `Tab`/`Shift+Tab` events, which tview Form handles natively for field-to-field movement.

---

## What is not yet built

- Actually connecting to MySQL (the `onConnect` callback in `main.go` is a placeholder modal)
- Database/table browser
- Query runner
- Copy/export functionality
