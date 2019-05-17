package main

import (
	"bytes"
)

const (
	CHUNK_SIZE = 4 * 1024
)

type Chunker interface {
	IndexEvent(event *Event, ch chan *bytes.Buffer)
}

func GetChunker(dir string, optimized bool) Chunker {
	if optimized {
		return optimizedChunker{
			prefix: dir,
			index:  make(ChunkIndex),
		}
	}
	return simpleChunker{prefix: dir}
}
