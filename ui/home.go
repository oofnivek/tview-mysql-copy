package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tview-mysql-copy/config"
)

func ShowHome(app *tview.Application, prefs *config.Preferences, onManageConnections func(), onManagePresets func(), onToggleTheme func()) tview.Primitive {
	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle(" Main Menu ").SetTitleAlign(tview.AlignLeft)
	list.AddItem("Manage Connections", "", 0, nil)
	list.AddItem("Manage Presets", "", 0, nil)

	icon, label := "◑", "dark"
	if prefs.Theme == "light" {
		icon, label = "◐", "light"
	}
	header := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight)
	header.SetText(fmt.Sprintf(" %s [yellow]%s mode[white]   [grey]t=toggle theme  q=quit ", icon, label))

	status := tview.NewTextView().SetDynamicColors(true)
	status.SetText("  [grey]↑↓/jk=navigate  Enter=select  t=toggle theme  q=quit")

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		idx := list.GetCurrentItem()
		count := list.GetItemCount()

		switch event.Key() {
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
		switch idx {
		case 0:
			onManageConnections()
		case 1:
			onManagePresets()
		}
	})

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(list, 0, 1, true).
		AddItem(status, 1, 0, false)
}
