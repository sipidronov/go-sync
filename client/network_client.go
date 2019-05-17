package main

import (
	"bytes"
	"errors"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"github.com/sipidronov/go-sync/message"
	"time"
)

type NetworkClient struct {
	task_queue chan *Event
	socket     *zmq.Socket
	poller     *zmq.Poller
	chunker    Chunker
	remote     string
}

func NewNetworkClient(task_queue chan *Event, chunker Chunker, remote string) *NetworkClient {
	socket, err := zmq.NewSocket(zmq.DEALER)
	poller := zmq.NewPoller()

	if err != nil {
		fmt.Println("Socket failed: ", err)
		return nil
	}
	return &NetworkClient{
		task_queue: task_queue,
		socket:     socket,
		poller:     poller,
		chunker:    chunker,
		remote:     remote,
	}
}

func (client *NetworkClient) processChanges() {
	fmt.Println("Ready to send")
	// TODO: set response handler
	for task := range client.task_queue {
		msgChan := make(chan *bytes.Buffer)
		go client.chunker.IndexEvent(task, msgChan)

		for msg := range msgChan {
			_, err := client.socket.SendMessage(msg)
			if err != nil {
				fmt.Println("Message send failed: ", err)
			}
		}
		fmt.Println("Event processing completed: ", task)
	}
}

func (client *NetworkClient) waitServer() error {
	var msg message.SyncMessage
	msg.Type = message.Hello
	buf, err := msg.Serialize()
	if err != nil {
		fmt.Println("Preparing Hello failed: ", err)
		return errors.New("Connection setup failed")
	}
	for {
		client.socket.SendMessage(buf.Bytes())
		polled, err := client.poller.Poll(1000 * time.Millisecond)
		if err == nil && len(polled) > 0 {
			client.socket.RecvMessageBytes(0)
			fmt.Println("Server replied.")
			break
		}
	}

	return nil
}

func (client *NetworkClient) Run() error {
	fmt.Println("Connecting to: ", client.remote)
	err := client.socket.Connect(client.remote)
	if err != nil {
		return err
	}

	client.poller.Add(client.socket, zmq.POLLIN)

	err = client.waitServer()
	if err != nil {
		return err
	}

	go client.processChanges()

	return nil
}
