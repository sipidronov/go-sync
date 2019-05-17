package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestWatchCreate(t *testing.T) {
	ch := make(chan *Event)
	watcher, err := NewWatcher(ch)

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	dir, err := ioutil.TempDir("", "watcher_test")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	defer os.RemoveAll(dir)
	watcher.Watch(dir)

	filename := "/TestWatchCreate.file"
	os.Create(dir + filename)

	event := <-ch

	if event == nil {
		t.Error("Nil task received")
		t.Fail()
	}

	if event.Type != Modified {
		t.Error("Wrong type: ", event.Type)
		t.Fail()
	}

	if strings.Compare(filename, event.Path) != 0 {
		t.Error("Wrong path: ", event.Path)
		t.Fail()
	}

	watcher.Close()
	for event = range ch {
		t.Error("Some messages left in chan: ", event)
	}
}

func TestWatchMove(t *testing.T) {
	ch := make(chan *Event)
	watcher, err := NewWatcher(ch)

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	dir, err := ioutil.TempDir("", "watcher_test")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	defer os.RemoveAll(dir)
	watcher.Watch(dir)

	filename := "/TestWatchCreate.file"
	os.Create(dir + filename)

	event := <-ch

	os.Rename(dir+filename, dir+"/TestWatchMove.file")
	event = <-ch

	if event.Type != Modified {
		t.Error("Wrong type: ", event.Type)
		t.Fail()
	}
	event = <-ch
	if event.Type != Deleted {
		t.Error("Wrong type: ", event.Type)
		t.Fail()
	}

	watcher.Close()
	for event = range ch {
		t.Error("Some messages left in chan: ", event)
	}
}

func TestIgnoreAttrs(t *testing.T) {
	ch := make(chan *Event)
	watcher, err := NewWatcher(ch)

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	dir, err := ioutil.TempDir("", "watcher_test")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	defer os.RemoveAll(dir)
	err = watcher.Watch(dir)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	filename := "/TestIgnoreAttrs.file"
	os.Create(dir + filename)
	event := <-ch

	if event.Type != Modified {
		t.Error("Wrong event: ", event)
		t.Fail()
	}

	err = os.Chmod(dir+filename, 666)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	os.Remove(dir + filename)
	event = <-ch

	if event.Type != Deleted {
		t.Error("Wrong type: ", event.Type)
		t.Fail()
	}

	watcher.Close()
	for event = range ch {
		t.Error("Some messages left in chan: ", event)
	}
}
