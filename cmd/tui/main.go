package main

import (
	"github.com/rivo/tview"
)

func main() {

	rightView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)

	rightView.SetBorder(true).
		SetTitle(" right ").
		SetTitleAlign(tview.AlignLeft)

	rightView.SetText("[yellow]welcome[-]")

	leftView := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	leftView.SetBorder(true).
		SetTitle(" left ").
		SetTitleAlign(tview.AlignLeft)

	leftView.AddItem("test", "", 0, func() {
		rightView.SetText("[red]test[-]")
	})

	barView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	barView.SetText(" [green]test[-]")

	flex := tview.NewFlex().
		AddItem(leftView, 0, 1, true).
		AddItem(rightView, 0, 2, false)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(flex, 0, 1, true).
		AddItem(barView, 1, 0, false)

	app := tview.NewApplication()
	if err := app.SetRoot(mainLayout, true).Run(); err != nil {
		panic(err)
	}
}
