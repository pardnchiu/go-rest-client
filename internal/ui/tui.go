package ui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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
	Mu        sync.RWMutex
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
		SetTitle(" API ").
		SetTitleAlign(tview.AlignLeft)

	tui.RightView.SetBorder(true).
		SetTitle(" Info ").
		SetTitleAlign(tview.AlignLeft)

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
		case tcell.KeyLeft:
			tui.App.SetFocus(tui.LeftView)
			return nil
		case tcell.KeyRight:
			tui.App.SetFocus(tui.RightView)
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
	t.Mu.RLock()
	requests := t.Requests
	t.Mu.RUnlock()

	t.LeftView.Clear()
	for i, e := range requests {
		index := i
		t.LeftView.AddItem(e.Name, "", 0, func() {
			t.sendRequest(index)
		})
	}
	t.LeftView.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	t.LeftView.SetSelectedTextColor(tcell.ColorWhite)
	t.LeftView.SetHighlightFullLine(true)
	if len(requests) > 0 {
		t.showRequestDetail(requests[0])
	}

	t.LeftView.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		t.Mu.RLock()
		defer t.Mu.RUnlock()

		if index >= 0 && index < len(t.Requests) {
			t.showRequestDetail(t.Requests[index])
		}
	})
}

func (t *TUI) showRequestDetail(req *Request) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[yellow]%s %s[-]\n", req.Method, req.URL))
	for k, v := range req.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	if req.Body != "" {
		sb.WriteString(fmt.Sprintf("\n[white]%s[-]", req.Body))
	}
	t.RightView.SetText(sb.String())

	t.HintView.SetText(
		fmt.Sprintf("%s %s",
			req.Method, req.URL))
}

func (t *TUI) sendRequest(index int) {
	t.Mu.RLock()
	if index < 0 || index >= len(t.Requests) {
		t.Mu.RUnlock()
		return
	}
	req := t.Requests[index]
	t.Mu.RUnlock()

	t.App.SetFocus(t.RightView)
	t.RightView.SetText("[yellow]Sending Resuest[-]")
	t.HintView.SetText(
		fmt.Sprintf("[yellow]Sending: %s[-]", req.Name))

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		var bodyReader io.Reader
		if req.Body != "" {
			bodyReader = bytes.NewBufferString(req.Body)
		}

		httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
		if err != nil {
			t.App.QueueUpdateDraw(func() {
				t.RightView.SetText(fmt.Sprintf("[red]Error creating request: %v[-]", err))
			})
			return
		}

		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}

		client := &http.Client{
			Timeout: 120 * time.Second,
		}

		start := time.Now()
		resp, err := client.Do(httpReq)
		if err != nil {
			t.App.QueueUpdateDraw(func() {
				t.RightView.SetText(fmt.Sprintf("[red]Error: %v[-]", err))
				t.HintView.SetText(
					fmt.Sprintf("[red]%v[-]", err))
			})
			return
		}
		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		isSSE := strings.Contains(contentType, "text/event-stream")

		if isSSE {
			t.handleSSEResponse(resp, start, req.Method, req.URL)
		} else {
			t.handleResponse(resp, start, req.Method, req.URL)
		}
	}()
}

func responseHeader(resp *http.Response) (string, *strings.Builder) {
	color := "green"
	if resp.StatusCode >= 400 {
		color = "red"
	} else if resp.StatusCode >= 300 {
		color = "yellow"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s]%s[-]\n", color, resp.Status))
	sb.WriteString("[yellow]Headers:[-]\n")
	for key, values := range resp.Header {
		for _, value := range values {
			sb.WriteString(fmt.Sprintf("  %s: [grey]%s[-]\n", key, value))
		}
	}
	return color, &sb
}

func (t *TUI) handleResponse(resp *http.Response, start time.Time, method, url string) {
	bodyBytes, err := io.ReadAll(resp.Body)
	duration := time.Since(start)

	t.App.QueueUpdateDraw(func() {
		statusColor, sb := responseHeader(resp)
		sb.WriteString(fmt.Sprintf("  Duration: [blue]%v[-]\n\n", duration))
		sb.WriteString("[yellow]Body:[-]\n")
		if err != nil {
			sb.WriteString(fmt.Sprintf("  [red]failed read body: %v[-]", err))
		} else {
			bodyStr := string(bodyBytes)

			var v any
			if json.Unmarshal(bodyBytes, &v) == nil {
				if formatted, err := json.MarshalIndent(v, "  ", "  "); err == nil {
					bodyStr = string(formatted)
				}
			}

			if resp.StatusCode >= 400 {
				sb.WriteString(fmt.Sprintf("  [red]%s[-]", bodyStr))
			} else {
				sb.WriteString(fmt.Sprintf("  %s", bodyStr))
			}
		}

		t.RightView.Clear()
		t.RightView.SetText(sb.String())

		t.HintView.SetText(
			fmt.Sprintf("[%s](%d)[-] %s %s | %v",
				statusColor, resp.StatusCode, method, url, duration))
	})
}

func (t *TUI) handleSSEResponse(resp *http.Response, start time.Time, method, url string) {
	statusColor, sb := responseHeader(resp)
	header := sb.String()

	var newSb strings.Builder
	t.App.QueueUpdateDraw(func() {
		newSb.WriteString(header)
		newSb.WriteString("\n[yellow]Stream Data:[-]\n")
		t.RightView.SetText(newSb.String())
	})

	scanner := bufio.NewScanner(resp.Body)
	var events []string
	count := 0

	for scanner.Scan() {
		row := scanner.Text()
		events = append(events, row)
		count++

		t.App.QueueUpdateDraw(func() {
			newSb.Reset()
			newSb.WriteString(header)
			newSb.WriteString(fmt.Sprintf("  Duration: [blue]%v[-] | Events: %d\n\n", time.Since(start), count/2))
			newSb.WriteString("[yellow]Stream Data:[-]\n\n")

			displayEvents := events
			if len(displayEvents) > 100 {
				displayEvents = displayEvents[len(displayEvents)-100:]
			}
			newSb.WriteString(strings.Join(displayEvents, "\n"))

			t.RightView.Clear()
			t.RightView.SetText(newSb.String())
			t.RightView.ScrollToEnd()
		})
	}

	if err := scanner.Err(); err != nil {
		t.App.QueueUpdateDraw(func() {
			t.HintView.SetText(
				fmt.Sprintf("[red]Stream error: %v[-]", err))
		})
		return
	}

	duration := time.Since(start)
	t.App.QueueUpdateDraw(func() {
		t.HintView.SetText(
			fmt.Sprintf("[%s](%d)[-] %s %s | %v",
				statusColor, resp.StatusCode, method, url, duration))
	})
}
