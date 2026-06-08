package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tview-mysql-copy/config"
	"tview-mysql-copy/db"
)

func buildPresetForm(app *tview.Application, existing *config.Preset, allConns []config.Connection, onSave func(config.Preset), onCancel func()) tview.Primitive {
	// current selections
	var srcConn, dstConn *config.Connection
	var srcDB, srcTable, dstDB string
	var dstConns []config.Connection

	// pending preselect values consumed as each async list populates (edit mode)
	pendSrcDB, pendSrcTable, pendDstConn, pendDstDB := "", "", "", ""
	if existing != nil {
		pendSrcDB = existing.SrcDatabase
		pendSrcTable = existing.SrcTable
		pendDstConn = existing.DstConnection
		pendDstDB = existing.DstDatabase
	}

	// lists
	srcConnList := tview.NewList().ShowSecondaryText(false)
	srcDBList := tview.NewList().ShowSecondaryText(false)
	srcTableList := tview.NewList().ShowSecondaryText(false)
	dstConnList := tview.NewList().ShowSecondaryText(false)
	dstDBList := tview.NewList().ShowSecondaryText(false)

	srcConnList.SetBorder(true).SetTitle(" Connection ").SetTitleAlign(tview.AlignLeft)
	srcDBList.SetBorder(true).SetTitle(" Database ").SetTitleAlign(tview.AlignLeft)
	srcTableList.SetBorder(true).SetTitle(" Table ").SetTitleAlign(tview.AlignLeft)
	dstConnList.SetBorder(true).SetTitle(" Connection ").SetTitleAlign(tview.AlignLeft)
	dstDBList.SetBorder(true).SetTitle(" Database ").SetTitleAlign(tview.AlignLeft)

	status := tview.NewTextView().SetDynamicColors(true)

	const keys = "   [grey]↑↓/jk=navigate  Enter=select  Ctrl+S=save  Esc=cancel"

	updateStatus := func() {
		switch {
		case srcConn == nil:
			status.SetText("  select source connection" + keys)
		case srcDB == "":
			status.SetText(fmt.Sprintf("  %s: select source database", srcConn.Name) + keys)
		case srcTable == "":
			status.SetText(fmt.Sprintf("  %s/%s: select source table", srcConn.Name, srcDB) + keys)
		case dstConn == nil:
			status.SetText(fmt.Sprintf("  [yellow]%s/%s/%s[white] → select destination connection", srcConn.Name, srcDB, srcTable) + keys)
		case dstDB == "":
			status.SetText(fmt.Sprintf("  [yellow]%s/%s/%s[white] → %s: select destination database", srcConn.Name, srcDB, srcTable, dstConn.Name) + keys)
		default:
			status.SetText(fmt.Sprintf("  [green]✓[white] %s/%s/%s → %s/%s/%s", srcConn.Name, srcDB, srcTable, dstConn.Name, dstDB, srcTable) + keys)
		}
	}

	findItem := func(list *tview.List, value string) int {
		for i := 0; i < list.GetItemCount(); i++ {
			main, _ := list.GetItemText(i)
			if main == value {
				return i
			}
		}
		return -1
	}

	setLoading := func(list *tview.List, title string) {
		list.Clear()
		list.SetTitle(fmt.Sprintf(" %s ", title))
		list.AddItem("[grey]Loading...", "", 0, nil)
	}

	setErr := func(list *tview.List, err error) {
		list.Clear()
		list.AddItem(fmt.Sprintf("[red]%v", err), "", 0, nil)
	}

	rebuildDstConns := func() {
		dstConns = dstConns[:0]
		for i := range allConns {
			if srcConn == nil || allConns[i].Name != srcConn.Name {
				dstConns = append(dstConns, allConns[i])
			}
		}
		dstConnList.Clear()
		for _, c := range dstConns {
			dstConnList.AddItem(c.Name, "", 0, nil)
		}
		dstConn = nil
		dstDB = ""
		dstDBList.Clear()
	}

	// forward declarations so handlers can call each other (preselect chain)
	var onSrcDBSelected func(idx int, name string)
	var onSrcTableSelected func(name string)
	var onDstConnSelected func(idx int, name string)
	var onDstDBSelected func(name string)

	onSrcConnSelected := func(idx int) {
		c := allConns[idx]
		srcConn = &c
		srcDB, srcTable = "", ""
		srcDBList.Clear()
		srcTableList.Clear()
		rebuildDstConns()
		updateStatus()

		pend := pendSrcDB
		pendSrcDB = ""
		setLoading(srcDBList, "Database")
		app.SetFocus(srcDBList)
		go func() {
			dbs, err := db.ListDatabases(c)
			app.QueueUpdateDraw(func() {
				srcDBList.Clear()
				srcDBList.SetTitle(" Database ")
				if err != nil {
					setErr(srcDBList, err)
					return
				}
				for _, d := range dbs {
					srcDBList.AddItem(d, "", 0, nil)
				}
				if pend != "" {
					if i := findItem(srcDBList, pend); i >= 0 {
						srcDBList.SetCurrentItem(i)
						onSrcDBSelected(i, pend)
					}
				}
			})
		}()
	}

	onSrcDBSelected = func(_ int, name string) {
		srcDB = name
		srcTable = ""
		srcTableList.Clear()
		updateStatus()

		pend := pendSrcTable
		pendSrcTable = ""
		setLoading(srcTableList, "Table")
		app.SetFocus(srcTableList)
		c := *srcConn
		go func() {
			tables, err := db.ListTables(c, name)
			app.QueueUpdateDraw(func() {
				srcTableList.Clear()
				srcTableList.SetTitle(" Table ")
				if err != nil {
					setErr(srcTableList, err)
					return
				}
				for _, t := range tables {
					srcTableList.AddItem(t, "", 0, nil)
				}
				if pend != "" {
					if i := findItem(srcTableList, pend); i >= 0 {
						srcTableList.SetCurrentItem(i)
						onSrcTableSelected(pend)
					}
				}
			})
		}()
	}

	onSrcTableSelected = func(name string) {
		srcTable = name
		updateStatus()

		pend := pendDstConn
		pendDstConn = ""
		if pend != "" {
			for i, c := range dstConns {
				if c.Name == pend {
					dstConnList.SetCurrentItem(i)
					onDstConnSelected(i, pend)
					return
				}
			}
		}
		app.SetFocus(dstConnList)
	}

	onDstConnSelected = func(idx int, _ string) {
		c := dstConns[idx]
		dstConn = &c
		dstDB = ""
		dstDBList.Clear()
		updateStatus()

		pend := pendDstDB
		pendDstDB = ""
		setLoading(dstDBList, "Database")
		app.SetFocus(dstDBList)
		go func() {
			dbs, err := db.ListDatabases(c)
			app.QueueUpdateDraw(func() {
				dstDBList.Clear()
				dstDBList.SetTitle(" Database ")
				if err != nil {
					setErr(dstDBList, err)
					return
				}
				for _, d := range dbs {
					dstDBList.AddItem(d, "", 0, nil)
				}
				if pend != "" {
					if i := findItem(dstDBList, pend); i >= 0 {
						dstDBList.SetCurrentItem(i)
						onDstDBSelected(pend)
					}
				}
			})
		}()
	}

	onDstDBSelected = func(name string) {
		dstDB = name
		updateStatus()
	}

	// wire SetSelectedFunc
	srcConnList.SetSelectedFunc(func(i int, _, _ string, _ rune) { onSrcConnSelected(i) })
	srcDBList.SetSelectedFunc(func(i int, name, _ string, _ rune) { onSrcDBSelected(i, name) })
	srcTableList.SetSelectedFunc(func(_ int, name, _ string, _ rune) { onSrcTableSelected(name) })
	dstConnList.SetSelectedFunc(func(i int, name, _ string, _ rune) { onDstConnSelected(i, name) })
	dstDBList.SetSelectedFunc(func(_ int, name, _ string, _ rune) { onDstDBSelected(name) })

	// shared key handler: j/k nav + Ctrl+S save + Esc cancel
	makeKeys := func(list *tview.List) func(*tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			idx := list.GetCurrentItem()
			count := list.GetItemCount()
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
			}
			switch event.Key() {
			case tcell.KeyCtrlS:
				if srcConn != nil && srcDB != "" && srcTable != "" && dstConn != nil && dstDB != "" {
					onSave(config.Preset{
						SrcConnection: srcConn.Name,
						SrcDatabase:   srcDB,
						SrcTable:      srcTable,
						DstConnection: dstConn.Name,
						DstDatabase:   dstDB,
						DstTable:      srcTable,
					})
				}
				return nil
			case tcell.KeyEscape:
				onCancel()
				return nil
			}
			return event
		}
	}
	srcConnList.SetInputCapture(makeKeys(srcConnList))
	srcDBList.SetInputCapture(makeKeys(srcDBList))
	srcTableList.SetInputCapture(makeKeys(srcTableList))
	dstConnList.SetInputCapture(makeKeys(dstConnList))
	dstDBList.SetInputCapture(makeKeys(dstDBList))

	// populate source connection list
	for _, c := range allConns {
		srcConnList.AddItem(c.Name, "", 0, nil)
	}
	// populate initial dst connection list (all, before src is chosen)
	rebuildDstConns()

	// preselect for edit mode
	if existing != nil {
		for i, c := range allConns {
			if c.Name == existing.SrcConnection {
				srcConnList.SetCurrentItem(i)
				onSrcConnSelected(i)
				break
			}
		}
	}

	updateStatus()

	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(srcConnList, 0, 1, true).
		AddItem(srcDBList, 0, 1, false).
		AddItem(srcTableList, 0, 2, false)

	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(dstConnList, 0, 1, false).
		AddItem(dstDBList, 0, 3, false)

	header := tview.NewFlex().
		AddItem(tview.NewTextView().SetText(" Source"), 0, 1, false).
		AddItem(tview.NewTextView().SetText("Destination ").SetTextAlign(tview.AlignRight), 0, 1, false)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(tview.NewFlex().
			AddItem(leftPanel, 0, 1, true).
			AddItem(rightPanel, 0, 1, false), 0, 1, true).
		AddItem(status, 1, 0, false)
}
