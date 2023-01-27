package MazeGameServer

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type WebApp struct {
	Connection  *websocket.Conn
	MessageType int
	MazeHost    *MazeHost
}

func CreateWebApp(Connection *websocket.Conn, MessageType int, Maze *MazeHost) (*WebApp, error) {
	var out = &WebApp{
		Connection:  Connection,
		MessageType: MessageType,
		MazeHost:    Maze,
	}

	err := out.SendMessage("MAZE " + Maze.MazeJson)
	if err != nil {
		return nil, err
	}

	go out.HandleMessages()

	return out, nil
}

func (m WebApp) SendMessage(message string) error {
	fmt.Printf("Sent Message to client: %s\n", message)
	if err := m.Connection.WriteMessage(m.MessageType, []byte(message+"\n")); err != nil {
		return errors.New(fmt.Sprintf("Failed to send message to WebApp: %s", err.Error()))
	}

	return nil
}

func (m WebApp) HandleMessages() {
	// read a message
	_, messageContent, err := m.Connection.ReadMessage()
	if err != nil {
		fmt.Printf("Handle message stopped: %s\n", err.Error())
		m.Disconnected()
		return
	}

	// print out that message
	fmt.Println(string(messageContent))

	var clientMessage = string(messageContent)

	log.Println("Client message: ", clientMessage)

	m.HandleMessages()
}

func (m WebApp) Disconnected() error {
	fmt.Printf("Disconnect webapp from code: %s\n", m.MazeHost.Code)
	err := m.Connection.Close()
	if err != nil {
		return err
	}

	return nil
}

func (m WebApp) SetPlayerPosition(JSON string) error {
	err := m.SendMessage(fmt.Sprintf("PlayerPosition %s", JSON))
	if err != nil {
		return err
	}
	return nil
}

func (m WebApp) SetMaze(JSON string) error {
	err := m.SendMessage(fmt.Sprintf("Maze %s", JSON))
	if err != nil {
		return err
	}
	return nil
}
