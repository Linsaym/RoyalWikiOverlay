package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("")

	w.SetContent(widget.NewLabel("hello world"))
	w.SetPadded(false)
	w.Resize(fyne.NewSize(200, 100))

	w.ShowAndRun()
}
