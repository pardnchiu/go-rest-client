package main

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/pardnchiu/go-rest-client/internal/parser"
	"github.com/pardnchiu/go-rest-client/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: program {file}.http")
		os.Exit(1)
	}

	filePath := os.Args[1]

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "file '%s' does not exist\n", filePath)
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get file '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	tui := ui.NewTUI()
	tui.Watcher = watcher
	if err := watcher.Add(filePath); err != nil {
		panic(err)
	}

	if _, err := parser.ReadFile(tui, filePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse file '%s': %v\n", filePath, err)
		os.Exit(1)
	}
	go parser.WatchFile(tui)

	if err := tui.App.Run(); err != nil {
		panic(err)
	}
}
