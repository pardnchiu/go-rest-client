package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUI struct {
	app       *tview.Application
	pages     *tview.Pages
	leftView  *tview.List
	rightView *tview.TextView
	hintView  *tview.TextView
}

func main() {
	tui := NewTUI()
	tui.readFile("test.http")

	if err := tui.app.Run(); err != nil {
		panic(err)
	}
}

func NewTUI() *TUI {
	tui := TUI{
		app:       tview.NewApplication(),
		pages:     tview.NewPages(),
		leftView:  tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true),
		rightView: tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWordWrap(true),
		hintView:  tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft),
	}

	tui.leftView.SetBorder(true).SetBorder(true).
		SetTitle(" left ").
		SetTitleAlign(tview.AlignLeft)

	tui.leftView.AddItem("test", "", 0, func() {
		tui.rightView.SetText("[red]test[-]")
	})

	tui.rightView.SetBorder(true).
		SetTitle(" right ").
		SetTitleAlign(tview.AlignLeft)

	tui.hintView.SetText(" [green]test[-]")

	flex := tview.NewFlex().
		AddItem(tui.leftView, 0, 1, true).
		AddItem(tui.rightView, 0, 2, false)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(flex, 0, 1, true).
		AddItem(tui.hintView, 1, 0, false)

	tui.app.SetRoot(mainLayout, true)

	tui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyEsc:
			tui.app.Stop()
			return nil
		case tcell.KeyTab:
			if tui.app.GetFocus() == tui.leftView {
				tui.app.SetFocus(tui.rightView)
				return nil
			}
			tui.app.SetFocus(tui.leftView)
			return nil
		}
		return event
	})

	return &tui
}

type Request struct {
	Name    string
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

func (t *TUI) readFile(path string) ([]*Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	regex := regexp.MustCompile(`^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+(.+)$`)

	var rowLen int
	var body []string
	var inBody bool
	var requests []*Request
	var request *Request

	var sb strings.Builder
	var name string
	for scanner.Scan() {
		rowLen++
		row := scanner.Text()
		text := strings.TrimSpace(row)

		if text == "###" {
			if request != nil {
				if len(body) > 0 {
					request.Body = strings.Join(body, "\n")
				}
				requests = append(requests, request)
			}
			request = nil
			inBody = false
			body = nil
			continue
		}

		if !inBody && (text == "" || strings.HasPrefix(text, "//") || strings.HasPrefix(text, "--")) {
			continue
		}

		if strings.HasPrefix(text, "#") && !strings.HasPrefix(text, "###") {
			continue
		}

		if text, found := strings.CutPrefix(text, "###"); found {
			name = strings.TrimSpace(text)
			continue
		}

		if matches := regex.FindStringSubmatch(text); matches != nil {
			if request != nil && len(body) > 0 {
				request.Body = strings.Join(body, "\n")
				requests = append(requests, request)
			}

			request = &Request{
				Method:  matches[1],
				URL:     matches[2],
				Headers: make(map[string]string),
				Name:    fmt.Sprintf("%s (%s %s)", name, matches[1], matches[2]),
			}
			inBody = false
			body = nil
			name = ""
			continue
		}

		if request != nil && !inBody {
			if text == "{" || text == "[" {
				inBody = true
				body = append(body, row)
				continue
			}

			if text == "" {
				inBody = true
				continue
			}

			if strings.Contains(row, ":") && !strings.HasPrefix(text, "\"") {
				parts := strings.SplitN(row, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])

					if checkHeader(key) {
						request.Headers[key] = value
						continue
					}
				}
			}

			if text != "" {
				inBody = true
				body = append(body, row)
			}
		}

		if inBody {
			body = append(body, row)
		}
	}

	if request != nil {
		if len(body) > 0 {
			request.Body = strings.Join(body, "\n")
		}
		requests = append(requests, request)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	t.leftView.Clear()
	for _, e := range requests {
		t.leftView.AddItem(e.Name, "", 0, func() {
			t.rightView.SetText("[red]test[-]")
		})
	}
	t.rightView.SetText(sb.String())

	return requests, nil
}

func checkHeader(name string) bool {
	if name == "" {
		return false
	}

	for _, ch := range name {
		if ch == ' ' || ch == '"' || ch == '\'' || ch == '\t' || ch < 33 || ch > 126 {
			return false
		}
	}

	for i, ch := range name {
		isLetter := (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
		isDigit := ch >= '0' && ch <= '9'
		isHyphen := ch == '-'
		isUnderscore := ch == '_'

		if i == 0 && !isLetter {
			return false
		}

		if !isLetter && !isDigit && !isHyphen && !isUnderscore {
			return false
		}
	}

	return true
}
