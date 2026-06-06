package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var darkTheme = tview.Theme{
	PrimitiveBackgroundColor:   tcell.ColorBlack,
	ContrastBackgroundColor:    tcell.ColorBlue,
	BorderColor:                tcell.ColorWhite,
	TitleColor:                 tcell.ColorWhite,
	GraphicsColor:              tcell.ColorWhite,
	PrimaryTextColor:           tcell.ColorWhite,
	SecondaryTextColor:         tcell.ColorYellow,
	TertiaryTextColor:          tcell.ColorGreen,
	InverseTextColor:           tcell.ColorBlue,
	ContrastSecondaryTextColor: tcell.ColorDarkCyan,
}

var lightTheme = tview.Theme{
	PrimitiveBackgroundColor:   tcell.ColorWhite,
	ContrastBackgroundColor:    tcell.ColorLightBlue,
	BorderColor:                tcell.ColorBlack,
	TitleColor:                 tcell.ColorBlack,
	GraphicsColor:              tcell.ColorBlack,
	PrimaryTextColor:           tcell.ColorBlack,
	SecondaryTextColor:         tcell.ColorNavy,
	TertiaryTextColor:          tcell.ColorDarkGreen,
	InverseTextColor:           tcell.ColorWhite,
	ContrastSecondaryTextColor: tcell.ColorTeal,
}

// ApplyTheme sets the global tview style. Call app.ForceDraw() after to refresh.
func ApplyTheme(theme string) {
	if theme == "light" {
		tview.Styles = lightTheme
	} else {
		tview.Styles = darkTheme
	}
}
