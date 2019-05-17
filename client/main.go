package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
)

var (
	flagHelp        = flag.Bool("h", false, "print help and exit")
	flagOptimize    = flag.Bool("optimize", true, "do not send un-changed chunks")
	flagPath        = flag.String("localpath", "", "Local directory path to track changes")
	flagRemote      = flag.String("remote", "", "zmq-like server address to send changes to")
	flagInitialSync = flag.Bool("initial-sync", true, "perform initial files sync")
)

func initialSync(dir string, event_chan chan *Event) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("Error listing directory: ", err)
		return errors.New("Listing failed")
	}

	if len(files) > 0 {
		fmt.Println("Will sync: ")
		for _, f := range files {
			fmt.Println("Will sync: ", f.Name())
			event_chan <- &Event{Type: Modified, Path: "/" + f.Name()}
		}
	} else {
		fmt.Println("Empty directory: ", dir)
	}

	return nil
}

func main() {
	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	event_chan := make(chan *Event, 1000)
	watcher, err := NewWatcher(event_chan)
	if err != nil {
		fmt.Println("Watcher failed: ", err)
		os.Exit(-1)
	}

	path := strings.TrimRight(*flagPath, "/")
	chunker := GetChunker(path, *flagOptimize)
	syncClient := NewNetworkClient(event_chan, chunker, *flagRemote)
	if syncClient == nil {
		fmt.Println("NetworkClient failed: ", err)
		os.Exit(-1)
	}

	err = syncClient.Run()
	if err != nil {
		fmt.Println("NetworkClient run failed: ", err)
		os.Exit(-1)
	}
	watcher.Watch(*flagPath)

	if *flagInitialSync {
		initialSync(path, event_chan)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	sig := <-sigs

	fmt.Println("Signal rcvd: ", sig)
	watcher.Close()

	<-event_chan

	fmt.Println("Shutdown completed")
	os.Exit(0)
}
