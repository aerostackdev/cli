package pkg

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Observer struct {
	watcher *fsnotify.Watcher
	indexer *Indexer
}

func NewObserver(indexer *Indexer) (*Observer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Observer{
		watcher: watcher,
		indexer: indexer,
	}, nil
}

func (o *Observer) Start(root string) {
	go func() {
		for {
			select {
			case event, ok := <-o.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					// Debounce or just re-index immediate
					if isIndexable(event.Name) {
						log.Printf("File changed: %s. Re-indexing...", event.Name)
						if err := o.indexer.IndexFile(event.Name); err != nil {
							log.Printf("Error indexing file %s: %v", event.Name, err)
						}
					}
				}
			case err, ok := <-o.watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	// Recursive watch
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return o.watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error starting watcher: %v", err)
	}
}

func (o *Observer) Close() error {
	return o.watcher.Close()
}

func isIndexable(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".go" || ext == ".ts" || ext == ".tsx"
}
