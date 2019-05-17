package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sipidronov/go-sync/message"
	"os"
)

type FileWriter struct {
	prefix string
}

type SyncWriter interface {
	Sync(msg *message.SyncMessage, buf *bytes.Buffer) error
}

func NewSyncWriter(dir string) (SyncWriter, error) {
	return FileWriter{prefix: dir}, nil
}

func (w FileWriter) ensureExist(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		os.Create(filename)
	}
}

func (w FileWriter) Sync(msg *message.SyncMessage, buf *bytes.Buffer) error {
	fullPath := w.prefix + msg.Path
	switch msg.Type {
	case message.FileDelete:
		err := os.Remove(fullPath)
		if err != nil {
			fmt.Println("Remove file ", fullPath, " failed: ", err)
			return errors.New("Remove failed")
		}

		fmt.Println("Removed: ", fullPath)
	case message.FileTruncate:
		w.ensureExist(fullPath)
		err := os.Truncate(fullPath, int64(msg.Offset))
		if err != nil {
			fmt.Println("Truncate file ", fullPath, " failed: ", err)
			return errors.New("Truncate failed")
		}

		fmt.Println("Truncated: ", fullPath, " to size: ", msg.Offset)
	case message.FileChunk:
		w.ensureExist(fullPath)
		f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		defer f.Close()
		if err != nil {
			fmt.Println("OpenFile failed: ", err)
			return errors.New("Modify failed")
		}

		_, err = f.WriteAt(buf.Bytes(), int64(msg.Offset))
		if err != nil {
			fmt.Println("WriteAt failed: ", err)
			return errors.New("Modify failed")
		}

		fmt.Println("Changed: ", fullPath, " at offset: ", msg.Offset)
	}

	return nil
}
