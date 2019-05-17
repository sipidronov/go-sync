package main

import (
	"bytes"
	"fmt"
	"github.com/sipidronov/go-sync/message"
	"io"
	"os"
)

type simpleChunker struct {
	prefix string
}

func (chunker simpleChunker) IndexEvent(event *Event, ch chan *bytes.Buffer) {
	defer close(ch)
	switch event.Type {
	case Deleted:
		fmt.Println("Sending delete command")
		chunker.sendDelete(event.Path, ch)
		return
	case Modified:
		fmt.Println("Sending whole file")
		chunker.sendFile(event.Path, ch)
	default:
		fmt.Println("Unexpected event type: ", event.Type)
	}
}

func (chunker *simpleChunker) sendDelete(path string, ch chan *bytes.Buffer) {
	msg := message.SyncMessage{
		Type:   message.FileDelete,
		Offset: 0,
		Size:   0,
		Path:   path,
	}

	data, err := msg.Serialize()
	if err != nil {
		fmt.Println("Message serialize failed: ", err)
	} else {
		ch <- data
	}
}

func (chunker *simpleChunker) sendFile(path string, ch chan *bytes.Buffer) {
	fullPath := chunker.prefix + path
	file, err := os.Open(fullPath)
	if err != nil {
		fmt.Println("Requested file open failed: ", err)
		return
	}

	defer file.Close()

	var offset uint64
	offset = 0
	for {
		msg := message.SyncMessage{}
		msg.Path = path
		msg.Offset = offset

		buffer := make([]byte, CHUNK_SIZE)
		n, err := file.Read(buffer)
		if err == io.EOF {
			msg.Type = message.FileTruncate
			msg.Size = 0
			data, _ := msg.Serialize()
			fmt.Println("FileTrunkate to:", msg.Offset)
			ch <- data
			break
		}

		fmt.Println("Chunk of size: ", n)

		msg.Type = message.FileChunk
		msg.Offset = offset
		msg.Size = uint64(n)

		data, err := msg.Serialize()
		data.Write(buffer[:n])
		if err != nil {
			fmt.Println("Skipping failed chunk: ", err)
		} else {
			ch <- data
		}

		offset += uint64(n)
	}
}
