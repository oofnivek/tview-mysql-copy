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

func buildPresetList(pc *config.PresetConfig, list *tview.List, onSelect func(idx int)) {
	order := sortedPresetIndices(pc)
	list.Clear()
	for _, i := range order {
		p := pc.Presets[i]
		label := fmt.Sprintf("%s/%s  →  %s/%s", p.SrcConnection, p.SrcTable, p.DstConnection, p.DstTable)
		list.AddItem(label, "", 0, nil)
	}
	list.AddItem("[Add new preset]", "", 0, nil)
	list.SetChangedFunc(func(displayIdx int, _, _ string, _ rune) {
		if displayIdx < len(order) {
			onSelect(order[displayIdx])
		} else {
			onSelect(displayIdx)
		}
	})
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
	header.SetText(fmt.Sprintf(" %s [yellow]%s mode[white]   [grey]t=toggle theme  Esc=back  q=quit ", icon, label))

	status := tview.NewTextView().SetDynamicColors(true)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(list, 0, 1, true).
		AddItem(status, 1, 0, false)

	var order []int

	refresh := func() {
		buildPresetList(pc, list, func(realIdx int) {
			if realIdx < len(pc.Presets) {
				p := pc.Presets[realIdx]
				status.SetText(fmt.Sprintf(
					" [yellow]%s/%s → %s/%s[white]  src_db=%s  dst_db=%s   [grey]↑↓/jk=nav  g/G=top/end  e=edit  d=delete",
					p.SrcConnection, p.SrcTable, p.DstConnection, p.DstTable, p.SrcDatabase, p.DstDatabase,
				))
			} else {
				status.SetText("  [grey]↑↓/jk=navigate   Enter or a=add new preset")
			}
		})
		order = sortedPresetIndices(pc)
	}

	realIdx := func(displayIdx int) int {
		if displayIdx < len(order) {
			return order[displayIdx]
		}
		return displayIdx
	}

	connNames := func() []string {
		names := make([]string, len(cfg.Connections))
		for i, c := range cfg.Connections {
			names[i] = c.Name
		}
		return names
	}

	openForm := func(idx int) {
		var existing *config.Preset
		if idx >= 0 && idx < len(pc.Presets) {
			p := pc.Presets[idx]
			existing = &p
		}
		form := buildPresetForm(existing, connNames(), func(p config.Preset) {
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
				pc.Presets = append(pc.Presets[:ri], pc.Presets[ri+1:]...)
				_ = config.SavePresets(pc)
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
		if idx < len(pc.Presets) {
			openForm(realIdx(idx))
		} else {
			openForm(-1)
		}
	})

	refresh()

	pages = tview.NewPages().AddPage("list", flex, true, true)
	return pages
}
