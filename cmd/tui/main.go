package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pardnchiu/go-rest-client/internal/parser"
	"github.com/pardnchiu/go-rest-client/internal/ui"
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	tui := ui.NewTUI()
	tui.Watcher = watcher
	if err := watcher.Add("test.http"); err != nil {
		panic(err)
	}

	parser.ReadFile(tui, "test.http")
	go parser.WatchFile(tui)

	if err := tui.App.Run(); err != nil {
		panic(err)
	}
}
