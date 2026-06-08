# AGENTS.md

## What this is

A Go terminal UI application built with [tview](https://github.com/rivo/tview) for copying MySQL tables between connections. Connections, presets, and preferences are stored locally in `~/.mysql-copy/`.

---

## Project layout

```
tview-mysql-copy/
├── main.go              Entry point — loads config/prefs/presets, applies theme, wires pages
├── config/
│   ├── config.go        Connection list: load/save ~/.mysql-copy/config.json
│   ├── prefs.go         User preferences: load/save ~/.mysql-copy/prefs.json (theme)
│   └── presets.go       Copy presets: load/save ~/.mysql-copy/presets.json
├── db/
│   └── mysql.go         Live MySQL queries — ListDatabases, ListTables
└── ui/
    ├── home.go          Main menu screen
    ├── connections.go   Connection list screen — navigation, CRUD
    ├── form.go          Add/edit connection form
    ├── presets.go       Preset list screen — navigation, CRUD
    ├── preset_form.go   Add/edit preset — two-panel live picker
    └── theme.go         Dark and light tview.Theme definitions + ApplyTheme()
```

---

## Data files (`~/.mysql-copy/`)

| File | Purpose |
|------|---------|
| `config.json` | Saved MySQL connections (name, host, port, user, password, database) |
| `prefs.json` | User preferences — currently only `theme` (`"dark"` or `"light"`) |
| `presets.json` | Copy presets — flat JSON array (see schema below) |

All files are created automatically on first run with safe permissions (`0600` files, `0700` directory).

---

## Preset schema

`presets.json` is a **flat JSON array** (not a wrapped object):

```json
[
  {
    "id": "b36511ffbfb7256b",
    "src_connection": "getgo2-vehicles-prod",
    "src_database":   "vehicles",
    "src_table":      "car_asset_owner",
    "dst_connection": "localhost",
    "dst_database":   "vehicles",
    "dst_table":      "car_asset_owner"
  }
]
```

- `id` is a random 8-byte hex string generated on creation (`config.NewPresetID()`).
- `dst_table` always equals `src_table` — the destination table is dropped and recreated from the source DDL.
- Source DDL has `AUTO_INCREMENT` removed before being applied to the destination.
- The file is **never reordered on save** — insertion order is preserved. Sorting is display-only.

---

## Key design decisions

### Theme
- **`tview.Styles` is global** — `ApplyTheme()` in `ui/theme.go` overwrites it directly. After changing it, the entire UI must be rebuilt (`buildUI()` in `main.go`) and `app.SetRoot` called again.
- Theme toggle key `t` is wired on every screen.

### Pages and navigation
- `main.go` owns the top-level `*tview.Pages` and holds the named pages `"home"`, `"connections"`, `"presets"`.
- Each screen receives `onBack` / `onToggleTheme` callbacks; it does not navigate directly.
- The connections and presets screens each own their own inner `*tview.Pages` so a form can be pushed on top of the list without rebuilding.

### Form navigation
- `ui/form.go` and `ui/preset_form.go` intercept `↑`/`↓` and convert them to `Tab`/`Shift+Tab`, which tview handles natively for field movement.

### Preset list display
- The preset list is **sorted by `src_connection` → `src_database` → `src_table`** at render time (`sortedPresetIndices`).
- A stable index mapping (`order []int`) translates display positions back to real `pc.Presets` slice indices for edit/delete operations.

### Preset form — two-panel live picker
The new preset UI (`buildPresetForm` in `ui/preset_form.go`) is a side-by-side panel, not a text input form:

```
 Source                                      Destination
┌─ Connection ──────────┐  ┌─ Connection ──────────────────┐
│ getgo2-vehicles-prod  │  │ localhost                     │
│ localhost             │  │ getgo2-cms-users              │
└───────────────────────┘  └───────────────────────────────┘
┌─ Database ────────────┐  ┌─ Database ──────────────────── ┐
│ vehicles              │  │ vehicles                       │
└───────────────────────┘  └────────────────────────────────┘
┌─ Table ───────────────┐
│ car_asset_owner       │
└───────────────────────┘
  context hint   ↑↓/jk=navigate  Enter=select  Ctrl+S=save  Esc=cancel
```

- **Left panel (Source):** Connection → Database → Table. Each list loads live from MySQL after the previous selection.
- **Right panel (Destination):** Connection → Database only. The destination table is not selected — it is recreated from the source DDL.
- The **source connection is excluded** from the destination connection list.
- **Panel alignment:** right panel uses proportions `1:3` (Connection:Database) to match left panel's `1:1:2` (Connection:Database:Table), keeping the Database borders on both sides level.
- Async loads use `go func() { ... app.QueueUpdateDraw(...) }()` and show "Loading..." while fetching.
- System databases (`information_schema`, `performance_schema`, `mysql`, `sys`) are filtered out of all database lists.
- Edit mode pre-populates via a chained pending-selection mechanism (`pendSrcDB`, `pendSrcTable`, `pendDstConn`, `pendDstDB`) consumed one-by-one as each async list loads.

---

## UI conventions

### Key bindings (all list screens)
| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Navigate list |
| `g` / `G` | Jump to top / bottom |
| `Enter` | Select / open |
| `e` | Edit selected item |
| `d` | Delete selected item |
| `a` | Add new item |
| `t` | Toggle dark/light theme |
| `Esc` | Go back |
| `q` | Quit app |
| `Ctrl+S` | Save (preset form only) |

### Status bar
- **Key hints always appear in the bottom status bar** — never duplicated in the header.
- The status bar must be populated **immediately on screen load**, not deferred to the first navigation event. Always call the status update function explicitly after building a list (do not rely on `SetChangedFunc` firing on first render).
- Format: `context info   [grey]key=action  key=action`

---

## What is not yet built

- Actually executing a copy (the copy logic reading source DDL, stripping `AUTO_INCREMENT`, recreating the table at the destination, and copying rows)
- The `onConnect` callback in `main.go` is still a placeholder modal
