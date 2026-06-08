package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tview-mysql-copy/config"
)

func buildPresetForm(existing *config.Preset, connNames []string, onSave func(config.Preset), onCancel func()) tview.Primitive {
	defaults := config.Preset{}
	if existing != nil {
		defaults = *existing
	}

	srcIdx, dstIdx := 0, 0
	for i, n := range connNames {
		if n == defaults.SrcConnection {
			srcIdx = i
		}
		if n == defaults.DstConnection {
			dstIdx = i
		}
	}

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Preset Details ").SetTitleAlign(tview.AlignLeft)

	if len(connNames) > 0 {
		form.AddDropDown("Source Connection", connNames, srcIdx, nil)
	} else {
		form.AddInputField("Source Connection", defaults.SrcConnection, 30, nil, nil)
	}
	form.AddInputField("Source Database", defaults.SrcDatabase, 30, nil, nil)
	form.AddInputField("Source Table", defaults.SrcTable, 30, nil, nil)

	if len(connNames) > 0 {
		form.AddDropDown("Dest Connection", connNames, dstIdx, nil)
	} else {
		form.AddInputField("Dest Connection", defaults.DstConnection, 30, nil, nil)
	}
	form.AddInputField("Dest Database", defaults.DstDatabase, 30, nil, nil)
	form.AddInputField("Dest Table", defaults.DstTable, 30, nil, nil)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown:
			return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
		case tcell.KeyUp:
			return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
		}
		return event
	})

	form.AddButton("Save", func() {
		p := config.Preset{
			SrcDatabase: form.GetFormItemByLabel("Source Database").(*tview.InputField).GetText(),
			SrcTable:    form.GetFormItemByLabel("Source Table").(*tview.InputField).GetText(),
			DstDatabase: form.GetFormItemByLabel("Dest Database").(*tview.InputField).GetText(),
			DstTable:    form.GetFormItemByLabel("Dest Table").(*tview.InputField).GetText(),
		}
		if len(connNames) > 0 {
			_, src := form.GetFormItemByLabel("Source Connection").(*tview.DropDown).GetCurrentOption()
			_, dst := form.GetFormItemByLabel("Dest Connection").(*tview.DropDown).GetCurrentOption()
			p.SrcConnection = src
			p.DstConnection = dst
		} else {
			p.SrcConnection = form.GetFormItemByLabel("Source Connection").(*tview.InputField).GetText()
			p.DstConnection = form.GetFormItemByLabel("Dest Connection").(*tview.InputField).GetText()
		}
		onSave(p)
	})

	form.AddButton("Cancel", onCancel)

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 0, 2, true).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)
}
