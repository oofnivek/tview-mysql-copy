package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tview-mysql-copy/config"
)

func buildConnectionList(cfg *config.Config, list *tview.List, onSelect func(idx int)) {
	list.Clear()
	for _, c := range cfg.Connections {
		label := fmt.Sprintf("%s  (%s@%s:%s/%s)", c.Name, c.User, c.Host, c.Port, c.Database)
		list.AddItem(label, "", 0, nil)
	}
	list.AddItem("[Add new connection]", "", 0, nil)
	list.SetChangedFunc(func(idx int, _, _ string, _ rune) {
		onSelect(idx)
	})
}

func ShowConnectionManager(app *tview.Application, cfg *config.Config, prefs *config.Preferences, onBack func(), onToggleTheme func(), onConnect func(c config.Connection)) tview.Primitive {
	var pages *tview.Pages

	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle(" Saved Connections ").SetTitleAlign(tview.AlignLeft)

	icon, label := "◑", "dark"
	if prefs.Theme == "light" {
		icon, label = "◐", "light"
	}
	header := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight)
	header.SetText(fmt.Sprintf(" %s [yellow]%s mode[white]   [grey]t=toggle theme  Esc=back  q=quit ", icon, label))

	status := tview.NewTextView().SetDynamicColors(true)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(list, 0, 1, true).
		AddItem(status, 1, 0, false)

	refresh := func() {
		buildConnectionList(cfg, list, func(idx int) {
			if idx < len(cfg.Connections) {
				c := cfg.Connections[idx]
				status.SetText(fmt.Sprintf(
					" [yellow]%s[white]  host=%s port=%s user=%s db=%s   [grey]↑↓/jk=nav  g/G=top/end  Enter=connect  e=edit  d=delete",
					c.Name, c.Host, c.Port, c.User, c.Database,
				))
			} else {
				status.SetText("  [grey]↑↓/jk=navigate   Enter or a=add new connection")
			}
		})
	}

	openForm := func(idx int) {
		var existing *config.Connection
		if idx >= 0 && idx < len(cfg.Connections) {
			c := cfg.Connections[idx]
			existing = &c
		}
		form := buildConnectionForm(existing, func(c config.Connection) {
			if idx >= 0 && idx < len(cfg.Connections) {
				cfg.Connections[idx] = c
			} else {
				cfg.Connections = append(cfg.Connections, c)
			}
			if err := config.Save(cfg); err != nil {
				status.SetText(fmt.Sprintf("[red]Save error: %v", err))
			}
			pages.SwitchToPage("list")
			refresh()
		}, func() {
			pages.SwitchToPage("list")
		})
		pages.AddAndSwitchToPage("form", form, true)
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		idx := list.GetCurrentItem()
		count := list.GetItemCount()

		switch event.Key() {
		case tcell.KeyEscape:
			onBack()
			return nil
		case tcell.KeyUp:
			if idx > 0 {
				list.SetCurrentItem(idx - 1)
			}
			return nil
		case tcell.KeyDown:
			if idx < count-1 {
				list.SetCurrentItem(idx + 1)
			}
			return nil
		}

		switch event.Rune() {
		case 'j':
			if idx < count-1 {
				list.SetCurrentItem(idx + 1)
			}
			return nil
		case 'k':
			if idx > 0 {
				list.SetCurrentItem(idx - 1)
			}
			return nil
		case 'g':
			list.SetCurrentItem(0)
			return nil
		case 'G':
			list.SetCurrentItem(count - 1)
			return nil
		case 'e':
			if idx < len(cfg.Connections) {
				openForm(idx)
				return nil
			}
		case 'd':
			if idx < len(cfg.Connections) {
				cfg.Connections = append(cfg.Connections[:idx], cfg.Connections[idx+1:]...)
				_ = config.Save(cfg)
				refresh()
				return nil
			}
		case 'a':
			openForm(-1)
			return nil
		case 't':
			onToggleTheme()
			return nil
		case 'q':
			app.Stop()
			return nil
		}
		return event
	})

	list.SetSelectedFunc(func(idx int, _, _ string, _ rune) {
		if idx < len(cfg.Connections) {
			onConnect(cfg.Connections[idx])
		} else {
			openForm(-1)
		}
	})

	refresh()

	pages = tview.NewPages().AddPage("list", flex, true, true)
	return pages
}
