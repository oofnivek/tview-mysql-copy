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

	ui.ApplyTheme(prefs.Theme)

	app := tview.NewApplication()

	app.SetRoot(ui.ShowConnectionManager(app, cfg, prefs, func(c config.Connection) {
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Connecting to [yellow]%s[white]\n%s@%s:%s/%s", c.Name, c.User, c.Host, c.Port, c.Database)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(_ int, _ string) {
				app.Stop()
			})
		app.SetRoot(modal, false)
	}), true)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
