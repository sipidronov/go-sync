package main

import (
	"bytes"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"github.com/sipidronov/go-sync/message"
	"time"
)

const (
	POLL_INTERVAL = 500
)

type NetworkServer struct {
	socket *zmq.Socket
	writer SyncWriter
}

func NewServer(endpoint string, writer SyncWriter) (*NetworkServer, error) {
	var server NetworkServer
	socket, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		return nil, err
	}

	err = socket.Bind(endpoint)
	if err != nil {
		return nil, err
	}

	server.socket = socket
	server.writer = writer

	return &server, nil
}

func (server *NetworkServer) Serve() {
	poller := zmq.NewPoller()
	poller.Add(server.socket, zmq.POLLIN)

	for {
		polled, err := poller.Poll(POLL_INTERVAL * time.Millisecond)

		if err == nil && len(polled) > 0 {
			msg, err := server.socket.RecvMessageBytes(0)
			if err != nil {
				fmt.Println("RecvMessage error: ", err)
				continue
			}

			// TODO: Reply with ack ot error back
			//identity := string(msg[0])
			buffer := bytes.NewBuffer(msg[1])
			sync_msg, err := message.DeserializeMessage(buffer)
			if err != nil {
				fmt.Println("DeserializeMessage failed: ", err)
				continue
			}

			if sync_msg.Type == message.Hello {
				fmt.Println("Echo back to client")
				server.socket.SendMessage(msg)
				continue
			}

			err = server.writer.Sync(sync_msg, buffer)
			if err != nil {
				fmt.Println("Sync failed: ", err)
				continue
			}
		}
	}
}
