package ui

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tview-mysql-copy/config"
)

func sortedPresetIndices(pc *config.PresetConfig) []int {
	indices := make([]int, len(pc.Presets))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(a, b int) bool {
		pa, pb := pc.Presets[indices[a]], pc.Presets[indices[b]]
		if pa.SrcConnection != pb.SrcConnection {
			return pa.SrcConnection < pb.SrcConnection
		}
		if pa.SrcDatabase != pb.SrcDatabase {
			return pa.SrcDatabase < pb.SrcDatabase
		}
		return pa.SrcTable < pb.SrcTable
	})
	return indices
}

func buildPresetList(pc *config.PresetConfig, list *tview.List) []int {
	order := sortedPresetIndices(pc)
	list.Clear()
	for _, i := range order {
		p := pc.Presets[i]
		list.AddItem(fmt.Sprintf("%s/%s  →  %s/%s", p.SrcConnection, p.SrcTable, p.DstConnection, p.DstTable), "", 0, nil)
	}
	return order
}

func ShowPresetManager(app *tview.Application, pc *config.PresetConfig, cfg *config.Config, prefs *config.Preferences, onBack func(), onToggleTheme func()) tview.Primitive {
	var pages *tview.Pages

	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle(" Presets ").SetTitleAlign(tview.AlignLeft)

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
		order = buildPresetList(pc, list)
	}

	openForm := func(idx int) {
		var existing *config.Preset
		if idx >= 0 && idx < len(pc.Presets) {
			p := pc.Presets[idx]
			existing = &p
		}
		form := buildPresetForm(app, existing, cfg.Connections, func(p config.Preset) {
			if idx >= 0 && idx < len(pc.Presets) {
				p.ID = pc.Presets[idx].ID
				pc.Presets[idx] = p
			} else {
				p.ID = config.NewPresetID()
				pc.Presets = append(pc.Presets, p)
			}
			if err := config.SavePresets(pc); err != nil {
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
			if idx < len(pc.Presets) {
				openForm(realIdx(idx))
				return nil
			}
		case 'd':
			if idx < len(pc.Presets) {
				ri := realIdx(idx)
				p := pc.Presets[ri]
				label := fmt.Sprintf("%s/%s → %s/%s", p.SrcConnection, p.SrcTable, p.DstConnection, p.DstTable)
				modal := tview.NewModal().
					SetText(fmt.Sprintf("Delete [yellow]%s[white]?", label)).
					AddButtons([]string{"Delete", "Cancel"}).
					SetDoneFunc(func(_ int, btn string) {
						pages.RemovePage("confirm")
						if btn == "Delete" {
							pc.Presets = append(pc.Presets[:ri], pc.Presets[ri+1:]...)
							_ = config.SavePresets(pc)
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

	list.SetSelectedFunc(func(_ int, _, _ string, _ rune) {})

	refresh()

	pages = tview.NewPages().AddPage("list", flex, true, true)
	return pages
}
