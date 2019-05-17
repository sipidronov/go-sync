package message

import (
	"strings"
	"testing"
)

func TestSerialize(t *testing.T) {
	msg := &SyncMessage{Type: FileChunk, Size: 1, Offset: 2, Path: "/tmp/foo.bar"}

	_, err := msg.Serialize()

	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestEndToEnd(t *testing.T) {
	msg := &SyncMessage{Type: FileChunk, Size: 1, Offset: 2, Path: "/tmp/foo.bar"}

	buf, err := msg.Serialize()
	if err != nil || buf == nil {
		t.Error("Serialize error")
		t.Fail()
	}

	buf.Write([]byte("SomeDataAfterMessage"))

	msg, err = DeserializeMessage(buf)

	if err != nil {
		t.Error("Deserialize error", err)
		t.Fail()
	}

	if msg == nil || msg.Type != FileChunk {
		t.Error("Invalid deserialize result")
		t.Fail()
	}

	if msg.Offset != 2 {
		t.Error("Invalid offset")
		t.Fail()
	}

	if strings.Compare(msg.Path, "/tmp/foo.bar") != 0 {
		t.Error("Path decoding failed. Result: ", msg.Path)
		t.Fail()
	}
}
