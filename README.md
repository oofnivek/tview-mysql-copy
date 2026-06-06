# tview-mysql-copy

A terminal UI for managing MySQL connections, built with Go and [tview](https://github.com/rivo/tview).

## Requirements

- Go 1.21+
- A terminal with colour support

## Install & run

```bash
git clone <repo>
cd tview-mysql-copy
go run .
```

Or build a binary:

```bash
go build -o tview-mysql-copy .
./tview-mysql-copy
```

## Connection manager

When you launch the app you land on the connection list.

```
 ◑ dark mode   t=toggle theme  q=quit
┌ Saved Connections ──────────────────────────────────┐
│  getgo2-vehicles-prod  (proddev@host:3306/)          │
│  localhost  (root@localhost:3306/)                   │
│  [Add new connection]                                │
└──────────────────────────────────────────────────────┘
 ↑↓/jk=nav  g/G=top/end  Enter=connect  e=edit  d=delete
```

### Keys

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `g` | Jump to first item |
| `G` | Jump to last item |
| `Enter` | Connect to selected connection |
| `e` | Edit selected connection |
| `d` | Delete selected connection |
| `a` / `Enter` on Add | Open add-connection form |
| `t` | Toggle dark / light theme |
| `q` | Quit |

## Add / edit a connection

Selecting **[Add new connection]** or pressing `e` opens a form:

```
┌ Connection Details ─────────────────────┐
│  Name      [ my-server                ] │
│  Host      [ 127.0.0.1               ] │
│  Port      [ 3306     ]               │
│  User      [ root                     ] │
│  Password  [ ********                 ] │
│  Database  [ mydb                     ] │
│                                         │
│  [ Save ]  [ Cancel ]                   │
└─────────────────────────────────────────┘
```

### Form keys

| Key | Action |
|-----|--------|
| `↓` / `Tab` | Next field |
| `↑` / `Shift+Tab` | Previous field |
| `Enter` on Save | Save and return to list |
| `Enter` on Cancel | Discard and return to list |

## Theme

Press `t` at any time on the connection list to switch between **dark** and **light** mode. The preference is saved immediately to `~/.mysql-copy/prefs.json` and restored on the next launch.

## Config files

All data is stored in `~/.mysql-copy/`:

| File | Contents |
|------|----------|
| `config.json` | Saved connections |
| `prefs.json` | Preferences (theme) |
