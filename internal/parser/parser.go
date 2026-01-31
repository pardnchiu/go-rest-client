package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pardnchiu/go-rest-client/internal/ui"
)

func ReadFile(t *ui.TUI, path string) ([]*ui.Request, error) {
	t.Filepath = path

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	regex := regexp.MustCompile(`^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+(.+)$`)

	var body []string
	var inBody bool
	var requests []*ui.Request
	var request *ui.Request

	var name string
	for scanner.Scan() {
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

			request = &ui.Request{
				Method:  matches[1],
				URL:     matches[2],
				Headers: make(map[string]string),
				Name:    name,
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

	t.Mu.Lock()
	t.Requests = requests
	t.Mu.Unlock()

	t.UpdateLeftView()

	return requests, nil
}

func WatchFile(t *ui.TUI) {
	var (
		mu    sync.Mutex
		timer *time.Timer
	)

	for {
		select {
		case event, ok := <-t.Watcher.Events:
			if !ok {
				return
			}

			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				mu.Lock()
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(200*time.Millisecond, func() {
					ReloadFile(t)
				})
				mu.Unlock()
			}

		case err, ok := <-t.Watcher.Errors:
			if !ok {
				return
			}
			t.App.QueueUpdateDraw(func() {
				t.HintView.SetText(fmt.Sprintf("[red]Watch error: %v[-]", err))
			})
		}
	}
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

func ReloadFile(t *ui.TUI) {
	t.App.QueueUpdateDraw(func() {
		t.HintView.SetText(
			fmt.Sprintf("[yellow]Edited at %s[-]",
				time.Now().Format("15:04:05")))
	})

	requests, err := ReadFile(t, t.Filepath)
	if err != nil {
		t.App.QueueUpdateDraw(func() {
			t.HintView.SetText(
				fmt.Sprintf("[red]%v at %s[-]", err,
					time.Now().Format("15:04:05")))
		})
		return
	}

	t.Mu.Lock()
	t.Requests = requests
	t.Mu.Unlock()

	t.App.QueueUpdateDraw(func() {
		index := t.LeftView.GetCurrentItem()

		t.UpdateLeftView()

		t.Mu.RLock()
		if index >= 0 && index < len(t.Requests) {
			t.LeftView.SetCurrentItem(index)
		}
		t.Mu.RUnlock()

		t.HintView.SetText(
			fmt.Sprintf("[green]Reloaded at %s[-]",
				time.Now().Format("15:04:05")))
	})
}
