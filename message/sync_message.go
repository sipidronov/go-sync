package message

import (
	"bytes"
	"encoding/binary"
)

type MsgType uint8

type FileOpType uint8

const (
	Hello MsgType = iota
	FileChunk
	FileDelete
	FileTruncate
	Ack
	Error
)

var byteOrder = binary.BigEndian

type SyncMessage struct {
	Type   MsgType
	Offset uint64
	Size   uint64
	Path   string
}

func (req *SyncMessage) Serialize() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, byteOrder, req.Type)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, byteOrder, req.Offset)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, byteOrder, req.Size)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, byteOrder, uint16(len(req.Path)))
	if err != nil {
		return nil, err
	}

	_, err = buf.Write([]byte(req.Path))
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func DeserializeMessage(buf *bytes.Buffer) (*SyncMessage, error) {
	var msg SyncMessage
	err := binary.Read(buf, byteOrder, &msg.Type)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buf, byteOrder, &msg.Offset)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buf, byteOrder, &msg.Size)
	if err != nil {
		return nil, err
	}

	var pathLen uint16
	err = binary.Read(buf, byteOrder, &pathLen)
	if err != nil {
		return nil, err
	}

	path := make([]byte, pathLen)
	_, err = buf.Read(path)
	if err != nil {
		return nil, err
	}

	msg.Path = string(path)
	return &msg, nil
}
