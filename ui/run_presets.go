package ui

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"tview-mysql-copy/config"
	"tview-mysql-copy/db"
)

type runState int

const (
	runPending runState = iota
	runRunning
	runDone
	runError
)

type presetRun struct {
	preset config.Preset
	state  runState
	copied int
	errMsg string
}

func ShowRunPresets(app *tview.Application, pc *config.PresetConfig, cfg *config.Config, prefs *config.Preferences, onBack func(), onToggleTheme func()) tview.Primitive {
	var innerPages *tview.Pages

	order := sortedPresetIndices(pc)
	selected := make(map[int]bool)

	findConn := func(name string) (config.Connection, bool) {
		for _, c := range cfg.Connections {
			if c.Name == name {
				return c, true
			}
		}
		return config.Connection{}, false
	}

	// ── Header ────────────────────────────────────────────────────────────────

	icon, themeLabel := "◑", "dark"
	if prefs.Theme == "light" {
		icon, themeLabel = "◐", "light"
	}
	header := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight)
	header.SetText(fmt.Sprintf(" %s [yellow]%s mode[-]   [grey]t=toggle theme ", icon, themeLabel))

	// ── Select page ───────────────────────────────────────────────────────────

	selectList := tview.NewList().ShowSecondaryText(false)
	selectList.SetBorder(true).SetTitle(" Select Presets to Run ").SetTitleAlign(tview.AlignLeft)

	rebuildSelectList := func(keepIdx int) {
		selectList.Clear()
		for i, ri := range order {
			p := pc.Presets[ri]
			prefix := "[ ] "
			if selected[i] {
				prefix = "[✓] "
			}
			selectList.AddItem(fmt.Sprintf("%s%s/%s  →  %s/%s", prefix, p.SrcConnection, p.SrcTable, p.DstConnection, p.DstTable), "", 0, nil)
		}
		if keepIdx >= 0 && keepIdx < selectList.GetItemCount() {
			selectList.SetCurrentItem(keepIdx)
		}
	}
	rebuildSelectList(-1)

	selectStatus := tview.NewTextView().SetDynamicColors(true)
	selectStatus.SetText("  [grey]↑↓/jk=navigate  g/G=top/end  Space=toggle  a=select all  n=select none  Enter=proceed  Esc=back  q=quit")

	selectFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(selectList, 0, 1, true).
		AddItem(selectStatus, 1, 0, false)

	// ── Progress page builder ─────────────────────────────────────────────────

	buildProgressPage := func(runs []*presetRun, threads int) tview.Primitive {
		var mu sync.Mutex
		allDone := false
		finished := 0

		textView := tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetWrap(false)
		textView.SetBorder(true).SetTitle(" Running Presets ").SetTitleAlign(tview.AlignLeft)

		progressStatus := tview.NewTextView().SetDynamicColors(true)

		render := func() {
			var sb strings.Builder
			for _, r := range runs {
				switch r.state {
				case runPending:
					fmt.Fprintf(&sb, "[grey][ PENDING ][-]  %s/%s  →  %s/%s\n",
						r.preset.SrcConnection, r.preset.SrcTable, r.preset.DstConnection, r.preset.DstTable)
				case runRunning:
					fmt.Fprintf(&sb, "[yellow][ RUNNING ][-]  %s/%s  →  %s/%s  (%d rows)\n",
						r.preset.SrcConnection, r.preset.SrcTable, r.preset.DstConnection, r.preset.DstTable, r.copied)
				case runDone:
					fmt.Fprintf(&sb, "[green][  DONE   ][-]  %s/%s  →  %s/%s  (%d rows)\n",
						r.preset.SrcConnection, r.preset.SrcTable, r.preset.DstConnection, r.preset.DstTable, r.copied)
				case runError:
					fmt.Fprintf(&sb, "[red][  ERROR  ][-]  %s/%s  →  %s/%s  [red]%s[-]\n",
						r.preset.SrcConnection, r.preset.SrcTable, r.preset.DstConnection, r.preset.DstTable, r.errMsg)
				}
			}
			textView.SetText(sb.String())
			if allDone {
				progressStatus.SetText(fmt.Sprintf("  [green]All done.[-]  %d/%d finished.  [grey]Esc=back  q=quit", finished, len(runs)))
			} else {
				progressStatus.SetText(fmt.Sprintf("  [yellow]Running...[-]  %d/%d done", finished, len(runs)))
			}
		}

		progressFlex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(textView, 0, 1, true).
			AddItem(progressStatus, 1, 0, false)

		progressFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			mu.Lock()
			isDone := allDone
			mu.Unlock()
			switch event.Key() {
			case tcell.KeyEscape:
				if isDone {
					innerPages.RemovePage("progress")
					innerPages.RemovePage("threads")
					innerPages.SwitchToPage("select")
					app.SetFocus(selectList)
				}
				return nil
			}
			switch event.Rune() {
			case 'q':
				if isDone {
					app.Stop()
				}
				return nil
			}
			return event
		})

		sem := make(chan struct{}, threads)
		var wg sync.WaitGroup

		for _, r := range runs {
			wg.Add(1)
			go func(run *presetRun) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				srcConn, ok := findConn(run.preset.SrcConnection)
				if !ok {
					mu.Lock()
					run.state = runError
					run.errMsg = "source connection not found"
					finished++
					mu.Unlock()
					app.QueueUpdateDraw(func() { mu.Lock(); render(); mu.Unlock() })
					return
				}
				dstConn, ok := findConn(run.preset.DstConnection)
				if !ok {
					mu.Lock()
					run.state = runError
					run.errMsg = "destination connection not found"
					finished++
					mu.Unlock()
					app.QueueUpdateDraw(func() { mu.Lock(); render(); mu.Unlock() })
					return
				}

				mu.Lock()
				run.state = runRunning
				mu.Unlock()
				app.QueueUpdateDraw(func() { mu.Lock(); render(); mu.Unlock() })

				err := db.CopyTable(
					srcConn, dstConn,
					run.preset.SrcDatabase, run.preset.SrcTable,
					run.preset.DstDatabase, run.preset.DstTable,
					func(copied int) {
						mu.Lock()
						run.copied = copied
						mu.Unlock()
						app.QueueUpdateDraw(func() { mu.Lock(); render(); mu.Unlock() })
					},
				)

				mu.Lock()
				if err != nil {
					run.state = runError
					run.errMsg = err.Error()
				} else {
					run.state = runDone
				}
				finished++
				mu.Unlock()
				app.QueueUpdateDraw(func() { mu.Lock(); render(); mu.Unlock() })
			}(r)
		}

		go func() {
			wg.Wait()
			mu.Lock()
			allDone = true
			mu.Unlock()
			app.QueueUpdateDraw(func() { mu.Lock(); render(); mu.Unlock() })
		}()

		render()
		return progressFlex
	}

	// ── Thread count page builder ─────────────────────────────────────────────

	buildThreadsPage := func(runs []*presetRun) tview.Primitive {
		form := tview.NewForm()
		form.SetBorder(true).SetTitle(" Parallel Threads ").SetTitleAlign(tview.AlignLeft)
		form.AddInputField("Threads", "3", 10, func(_ string, lastChar rune) bool {
			return lastChar >= '0' && lastChar <= '9'
		}, nil)
		form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyDown:
				return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
			case tcell.KeyUp:
				return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
			case tcell.KeyEscape:
				innerPages.SwitchToPage("select")
				app.SetFocus(selectList)
				return nil
			}
			return event
		})
		form.AddButton("Run", func() {
			txt := form.GetFormItemByLabel("Threads").(*tview.InputField).GetText()
			n := 3
			if v, err := strconv.Atoi(txt); err == nil && v >= 1 {
				n = v
			}
			progress := buildProgressPage(runs, n)
			innerPages.AddAndSwitchToPage("progress", progress, true)
			app.SetFocus(progress)
		})
		form.AddButton("Cancel", func() {
			innerPages.SwitchToPage("select")
			app.SetFocus(selectList)
		})

		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(form, 9, 0, true).
				AddItem(nil, 0, 1, false), 40, 0, true).
			AddItem(nil, 0, 1, false)
	}

	// ── Select input capture ──────────────────────────────────────────────────

	selectList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		idx := selectList.GetCurrentItem()
		count := selectList.GetItemCount()

		switch event.Key() {
		case tcell.KeyEscape:
			onBack()
			return nil
		case tcell.KeyUp:
			if idx > 0 {
				selectList.SetCurrentItem(idx - 1)
			}
			return nil
		case tcell.KeyDown:
			if idx < count-1 {
				selectList.SetCurrentItem(idx + 1)
			}
			return nil
		case tcell.KeyEnter:
			var runs []*presetRun
			for i, ri := range order {
				if selected[i] {
					runs = append(runs, &presetRun{preset: pc.Presets[ri]})
				}
			}
			if len(runs) == 0 {
				return nil
			}
			tp := buildThreadsPage(runs)
			innerPages.AddAndSwitchToPage("threads", tp, true)
			return nil
		}

		switch event.Rune() {
		case ' ':
			selected[idx] = !selected[idx]
			rebuildSelectList(idx)
			return nil
		case 'a':
			for i := range order {
				selected[i] = true
			}
			rebuildSelectList(idx)
			return nil
		case 'n':
			for i := range selected {
				selected[i] = false
			}
			rebuildSelectList(idx)
			return nil
		case 'j':
			if idx < count-1 {
				selectList.SetCurrentItem(idx + 1)
			}
			return nil
		case 'k':
			if idx > 0 {
				selectList.SetCurrentItem(idx - 1)
			}
			return nil
		case 'g':
			selectList.SetCurrentItem(0)
			return nil
		case 'G':
			selectList.SetCurrentItem(count - 1)
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

	innerPages = tview.NewPages().AddPage("select", selectFlex, true, true)
	return innerPages
}
