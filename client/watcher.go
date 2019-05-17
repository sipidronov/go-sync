package main

import (
	"fmt"
	"github.com/howeyc/fsnotify"
	"strings"
)

type Watcher struct {
	iowatcher *fsnotify.Watcher
	eventChan chan *Event
	dir       string
}

func NewWatcher(ch chan *Event) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{iowatcher: watcher, eventChan: ch}, nil
}

func mapEventType(event *fsnotify.FileEvent) (ChangeType, error) {
	if event.IsAttrib() {
		return Unsupported, fmt.Errorf("Attrib change is not supported: %s", event)
	} else if event.IsCreate() || event.IsModify() {
		return Modified, nil
	} else if event.IsDelete() || event.IsRename() {
		return Deleted, nil
	} else {
		return Unsupported, fmt.Errorf("Un-supported event: %s", event)
	}
}

func (w *Watcher) handleEvent() {
	watcher := w.iowatcher
	defer fmt.Println("Handler stopped")
	defer close(w.eventChan)
	fmt.Println("Ready to watch")
	for {
		select {
		case event, ok := <-watcher.Event:
			if !ok {
				return
			}
			fmt.Println("event: ", event)
			relPath := strings.TrimPrefix(event.Name, w.dir)
			eventType, err := mapEventType(event)

			if err != nil {
				fmt.Println("Ignoring io event: ", err)
			} else {
				w.eventChan <- &Event{Type: eventType, Path: relPath}
				fmt.Println("Sent event type: ", eventType)
			}
		case err, ok := <-watcher.Error:
			if !ok {
				return
			}
			fmt.Println("error: ", err)
		}
	}
}

func (w *Watcher) Watch(path string) error {
	w.dir = path
	go w.handleEvent()
	return w.iowatcher.WatchFlags(path, fsnotify.FSN_ALL)
}

func (w *Watcher) Close() error {
	return w.iowatcher.Close()
}
