package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher *fsnotify.Watcher
	paths   []string
}

func NewFileWatcher() *FileWatcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	return &FileWatcher{
		watcher: watcher,
		paths:   make([]string, 0),
	}
}

func (fw *FileWatcher) AddPath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Add to watcher
	err := fw.watcher.Add(path)
	if err != nil {
		return fmt.Errorf("failed to add path to watcher: %w", err)
	}

	fw.paths = append(fw.paths, path)
	fmt.Printf("Now watching: %s\n", path)
	return nil
}

func (fw *FileWatcher) Start() {
	defer fw.watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-fw.watcher.Events:
				if !ok {
					return
				}

				timestamp := time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("[%s] %s: %s\n", timestamp, event.Op, event.Name)

				// Handle specific events
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					fmt.Printf("  → File modified: %s\n", filepath.Base(event.Name))
				case event.Op&fsnotify.Create == fsnotify.Create:
					fmt.Printf("  → File created: %s\n", filepath.Base(event.Name))
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					fmt.Printf("  → File removed: %s\n", filepath.Base(event.Name))
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					fmt.Printf("  → File renamed: %s\n", filepath.Base(event.Name))
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					fmt.Printf("  → File permissions changed: %s\n", filepath.Base(event.Name))
				}

			case err, ok := <-fw.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Error: %v", err)
			}
		}
	}()

	fmt.Println("File watcher started. Press Ctrl+C to stop.")
	<-done
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run file-watcher.go <path1> [path2] [path3] ...")
		fmt.Println("Example: go run file-watcher.go . /path/to/directory /path/to/file.txt")
		os.Exit(1)
	}

	watcher := NewFileWatcher()

	// Add all provided paths
	for _, path := range os.Args[1:] {
		if err := watcher.AddPath(path); err != nil {
			log.Printf("Failed to add path %s: %v", path, err)
		}
	}

	if len(watcher.paths) == 0 {
		fmt.Println("No valid paths to watch")
		os.Exit(1)
	}

	watcher.Start()
}
