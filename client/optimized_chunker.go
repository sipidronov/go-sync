package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/sipidronov/go-sync/message"
	"io"
	"os"
)

type ChunkIndex map[string][]*indexChunkEntry

type indexChunkEntry struct {
	Hash *[md5.Size]byte
}

type optimizedChunker struct {
	prefix string
	index  ChunkIndex
}

func (chunker optimizedChunker) IndexEvent(event *Event, ch chan *bytes.Buffer) {
	defer close(ch)
	switch event.Type {
	case Deleted:
		fmt.Println("Sending delete command")
		delete(chunker.index, event.Path)
		chunker.sendDelete(event.Path, ch)
		return
	case Modified:
		chunker.sendChanged(event.Path, ch)
	default:
		fmt.Println("Unexpected event type: ", event.Type)
	}
}

func (chunker *optimizedChunker) sendDelete(path string, ch chan *bytes.Buffer) {
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

func (chunker *optimizedChunker) indexEquals(entry *indexChunkEntry, newValue *[md5.Size]byte) bool {
	if bytes.Compare(entry.Hash[:], newValue[:]) == 0 {
		return true
	} else {
		return false
	}
}

func (chunker *optimizedChunker) getIndexEntry(key string, pos int) *indexChunkEntry {
	list, ok := chunker.index[key]
	if !ok || pos >= len(list) {
		return nil
	}

	return list[pos]
}

func (chunker *optimizedChunker) updateIndexEntry(key string, pos int, value *indexChunkEntry) error {
	_, ok := chunker.index[key]
	if !ok {
		chunker.index[key] = make([]*indexChunkEntry, 1)
	}

	chunkCount := len(chunker.index[key])

	if pos > chunkCount {
		fmt.Println("Desired position out of bound. Want: ", pos, " size: ", chunkCount)
		return errors.New("Index position out of bounds")
	} else if pos < chunkCount {
		chunker.index[key][pos] = value
	} else {
		chunker.index[key] = append(chunker.index[key], value)
	}

	return nil
}

func (chunker *optimizedChunker) trimIndexEntry(key string, pos int) error {
	list, ok := chunker.index[key]
	if !ok || pos > len(list) {
		return errors.New("Trim index out of bounds")
	}

	chunker.index[key] = list[:pos]

	return nil
}

func (chunker *optimizedChunker) sendTrim(path string, offset uint64, ch chan *bytes.Buffer) {
	msg := message.SyncMessage{}
	msg.Path = path
	msg.Offset = offset
	msg.Type = message.FileTruncate
	msg.Size = 0
	data, _ := msg.Serialize()
	fmt.Println("FileTruncate to:", msg.Offset)
	ch <- data
}

func (chunker *optimizedChunker) sendChanged(path string, ch chan *bytes.Buffer) {
	fullPath := chunker.prefix + path
	file, err := os.Open(fullPath)
	if err != nil {
		fmt.Println("Requested file open failed: ", err)
		return
	}

	defer file.Close()

	var offset uint64
	offset = 0
	chunkPos := 0
	for {
		buffer := make([]byte, CHUNK_SIZE)
		n, err := file.Read(buffer)
		if err == io.EOF {
			chunker.trimIndexEntry(path, chunkPos)
			chunker.sendTrim(path, offset, ch)
			break
		}

		fmt.Println("Chunk of size: ", n)

		hash := md5.Sum(buffer[:n])
		indexEntry := chunker.getIndexEntry(path, chunkPos)
		if indexEntry != nil && chunker.indexEquals(indexEntry, &hash) {
			fmt.Println("Chunk not changed at offset: ", offset)
		} else {
			chunker.updateIndexEntry(path, chunkPos, &indexChunkEntry{Hash: &hash})

			msg := message.SyncMessage{}
			msg.Type = message.FileChunk
			msg.Path = path
			msg.Offset = offset
			msg.Size = uint64(n)

			data, err := msg.Serialize()
			data.Write(buffer[:n])
			if err != nil {
				fmt.Println("Skipping failed chunk: ", err)
			} else {
				ch <- data
			}
		}
		offset += uint64(n)
		chunkPos += 1
	}
}
