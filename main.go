package main

import (
	"fmt"
	"log"

	"github.com/rivo/tview"
	"tview-mysql-copy/config"
	"tview-mysql-copy/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	prefs, err := config.LoadPrefs()
	if err != nil {
		log.Fatalf("failed to load preferences: %v", err)
	}

	presets, err := config.LoadPresets()
	if err != nil {
		log.Fatalf("failed to load presets: %v", err)
	}

	ui.ApplyTheme(prefs.Theme)

	app := tview.NewApplication()

	// pages is reassigned on each theme rebuild; closures capture the variable,
	// so they always reference the current pages instance.
	var pages *tview.Pages
	currentPage := "home"

	var buildUI func()

	toggleTheme := func() {
		if prefs.Theme == "light" {
			prefs.Theme = "dark"
		} else {
			prefs.Theme = "light"
		}
		_ = config.SavePrefs(prefs)
		ui.ApplyTheme(prefs.Theme)
		buildUI()
		if currentPage != "home" {
			pages.SwitchToPage(currentPage)
		}
		app.SetRoot(pages, true)
	}

	buildUI = func() {
		pages = tview.NewPages()

		connections := ui.ShowConnectionManager(app, cfg, prefs,
			func() { // onBack
				currentPage = "home"
				pages.SwitchToPage("home")
			},
			toggleTheme,
			func(c config.Connection) { // onConnect (placeholder)
				modal := tview.NewModal().
					SetText(fmt.Sprintf("Connecting to [yellow]%s[white]\n%s@%s:%s/%s", c.Name, c.User, c.Host, c.Port, c.Database)).
					AddButtons([]string{"OK"}).
					SetDoneFunc(func(_ int, _ string) {
						app.Stop()
					})
				app.SetRoot(modal, false)
			},
		)

		presetsPage := ui.ShowPresetManager(app, presets, cfg, prefs,
			func() { // onBack
				currentPage = "home"
				pages.SwitchToPage("home")
			},
			toggleTheme,
		)

		home := ui.ShowHome(app, prefs,
			func() { // onManageConnections
				currentPage = "connections"
				pages.SwitchToPage("connections")
			},
			func() { // onManagePresets
				currentPage = "presets"
				pages.SwitchToPage("presets")
			},
			toggleTheme,
		)

		pages.AddPage("home", home, true, true)
		pages.AddPage("connections", connections, true, false)
		pages.AddPage("presets", presetsPage, true, false)
	}

	buildUI()
	app.SetRoot(pages, true)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
