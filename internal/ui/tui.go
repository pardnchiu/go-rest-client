package ui

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUI struct {
	App       *tview.Application
	Pages     *tview.Pages
	LeftView  *tview.List
	RightView *tview.TextView
	HintView  *tview.TextView
	Watcher   *fsnotify.Watcher
	Filepath  string
	Requests  []*Request
}

type Request struct {
	Name    string
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

func NewTUI() *TUI {
	tui := TUI{
		App:       tview.NewApplication(),
		Pages:     tview.NewPages(),
		LeftView:  tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true),
		RightView: tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWordWrap(true),
		HintView:  tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft),
	}

	tui.LeftView.SetBorder(true).SetBorder(true).
		SetTitle(" left ").
		SetTitleAlign(tview.AlignLeft)

	tui.LeftView.AddItem("test", "", 0, func() {
		tui.RightView.SetText("[red]test[-]")
	})

	tui.RightView.SetBorder(true).
		SetTitle(" right ").
		SetTitleAlign(tview.AlignLeft)

	tui.HintView.SetText(" [green]test[-]")

	flex := tview.NewFlex().
		AddItem(tui.LeftView, 0, 1, true).
		AddItem(tui.RightView, 0, 2, false)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(flex, 0, 1, true).
		AddItem(tui.HintView, 1, 0, false)

	tui.App.SetRoot(mainLayout, true)

	tui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyEsc:
			tui.App.Stop()
			return nil
		case tcell.KeyTab:
			if tui.App.GetFocus() == tui.LeftView {
				tui.App.SetFocus(tui.RightView)
				return nil
			}
			tui.App.SetFocus(tui.LeftView)
			return nil
		}
		return event
	})

	return &tui
}

func (t *TUI) UpdateLeftView() {
	t.LeftView.Clear()
	for _, e := range t.Requests {
		t.LeftView.AddItem(e.Name, "", 0, func() {
			t.RightView.SetText("[red]test[-]")
		})
	}
}

func (t *TUI) Info(message string) {
	t.HintView.SetText(fmt.Sprintf("[yellow]✱ %s at %s[-]", fmt.Sprint(message), time.Now().Format("15:04:05")))
}

func (t *TUI) Okay(message string) {
	t.HintView.SetText(fmt.Sprintf("[green]✓ %s at %s[-]", fmt.Sprint(message), time.Now().Format("15:04:05")))
}

func (t *TUI) Error(err error) {
	t.HintView.SetText(fmt.Sprintf("[red]✕ %v at %s[-]", err, time.Now().Format("15:04:05")))
}
