package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	flagHelp   = flag.Bool("h", false, "print help and exit")
	flagPath   = flag.String("localpath", "/tmp", "Local directory path write changes to")
	flagRemote = flag.String("endpoint", "", "zmq-like endpoint to listen on")
)

func main() {
	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	writer, err := NewSyncWriter(strings.TrimRight(*flagPath, "/"))
	if err != nil {
		fmt.Println("NewSyncWriter failed: ", err)
		os.Exit(-1)
	}

	server, err := NewServer(*flagRemote, writer)
	if err != nil {
		fmt.Println("NewServer failed: ", err)
		os.Exit(-1)
	}

	server.Serve()
}
