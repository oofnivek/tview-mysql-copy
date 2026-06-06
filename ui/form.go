package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tview-mysql-copy/config"
)

func buildConnectionForm(existing *config.Connection, onSave func(config.Connection), onCancel func()) tview.Primitive {
	defaults := config.Connection{Port: "3306"}
	if existing != nil {
		defaults = *existing
	}

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Connection Details ").SetTitleAlign(tview.AlignLeft)

	form.AddInputField("Name",        defaults.Name,     30, nil, nil)
	form.AddInputField("Host",        defaults.Host,     30, nil, nil)
	form.AddInputField("Port",        defaults.Port,     10, nil, nil)
	form.AddInputField("User",        defaults.User,     30, nil, nil)
	form.AddPasswordField("Password", defaults.Password, 30, '*', nil)
	form.AddInputField("Database",    defaults.Database, 30, nil, nil)

	// Up/Down arrows navigate between fields (Tab/Shift+Tab already work natively).
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
		c := config.Connection{
			Name:     form.GetFormItemByLabel("Name").(*tview.InputField).GetText(),
			Host:     form.GetFormItemByLabel("Host").(*tview.InputField).GetText(),
			Port:     form.GetFormItemByLabel("Port").(*tview.InputField).GetText(),
			User:     form.GetFormItemByLabel("User").(*tview.InputField).GetText(),
			Password: form.GetFormItemByLabel("Password").(*tview.InputField).GetText(),
			Database: form.GetFormItemByLabel("Database").(*tview.InputField).GetText(),
		}
		onSave(c)
	})

	form.AddButton("Cancel", onCancel)

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 0, 2, true).
			AddItem(nil, 0, 1, false), 50, 0, true).
		AddItem(nil, 0, 1, false)
}
