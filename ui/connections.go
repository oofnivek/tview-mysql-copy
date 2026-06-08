package ui

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tview-mysql-copy/config"
)

func sortedConnIndices(cfg *config.Config) []int {
	indices := make([]int, len(cfg.Connections))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(a, b int) bool {
		return cfg.Connections[indices[a]].Name < cfg.Connections[indices[b]].Name
	})
	return indices
}

func buildConnectionList(cfg *config.Config, list *tview.List) []int {
	order := sortedConnIndices(cfg)
	list.Clear()
	for _, i := range order {
		c := cfg.Connections[i]
		list.AddItem(fmt.Sprintf("%s  (%s@%s:%s/%s)", c.Name, c.User, c.Host, c.Port, c.Database), "", 0, nil)
	}
	return order
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
	header.SetText(fmt.Sprintf(" %s [yellow]%s mode[white]   [grey]t=toggle theme ", icon, label))

	status := tview.NewTextView().SetDynamicColors(true)
	status.SetText("  [grey]↑↓/jk=navigate  g/G=top/end  a=add  e=edit  d=delete  Esc=back  q=quit")

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(list, 0, 1, true).
		AddItem(status, 1, 0, false)

	var order []int

	realIdx := func(displayIdx int) int {
		if displayIdx < len(order) {
			return order[displayIdx]
		}
		return displayIdx
	}

	refresh := func() {
		order = buildConnectionList(cfg, list)
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
				openForm(realIdx(idx))
				return nil
			}
		case 'd':
			if idx < len(cfg.Connections) {
				ri := realIdx(idx)
				name := cfg.Connections[ri].Name
				modal := tview.NewModal().
					SetText(fmt.Sprintf("Delete [yellow]%s[white]?", name)).
					AddButtons([]string{"Delete", "Cancel"}).
					SetDoneFunc(func(_ int, label string) {
						pages.RemovePage("confirm")
						app.SetFocus(list)
						if label == "Delete" {
							cfg.Connections = append(cfg.Connections[:ri], cfg.Connections[ri+1:]...)
							_ = config.Save(cfg)
							refresh()
						}
					})
				pages.AddAndSwitchToPage("confirm", modal, false)
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


	refresh()

	pages = tview.NewPages().AddPage("list", flex, true, true)
	return pages
}
