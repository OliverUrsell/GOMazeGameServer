package MazeGameServer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type WebApp struct {
	Connection  *websocket.Conn
	MessageType int
	MazeHost    *MazeHost
	IsGuide     bool // If not a guide they're a monster controller
}

func CreateWebApp(Connection *websocket.Conn, MessageType int, Maze *MazeHost) (*WebApp, error) {
	var out = &WebApp{
		Connection:  Connection,
		MessageType: MessageType,
		MazeHost:    Maze,
		IsGuide:     true,
	}

	// Add whether this player is a guide or monster to the json
	var mazeJSON map[string]interface{}

	err := json.Unmarshal([]byte(Maze.MazeJson), &mazeJSON)
	if err != nil {
		return nil, err
	}

	if Maze.MonsterControllerConnected {
		mazeJSON["player_type"] = "guide"
	} else {
		mazeJSON["player_type"] = "monster"
		Maze.MonsterControllerConnected = true
		out.IsGuide = false
	}

	jsonBytes, err := json.Marshal(mazeJSON)
	if err != nil {
		return nil, err
	}

	err = out.SendMessage("MAZE " + string(jsonBytes))
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
	fmt.Printf("Disconnected webapp from code: %s\n", m.MazeHost.Code)
	err := m.Connection.Close()
	if err != nil {
		return err
	}

	if !m.IsGuide {
		m.MazeHost.MonsterControllerConnected = false
	}

	return nil
}

func (m WebApp) SetPositions(JSON string) error {
	err := m.SendMessage(fmt.Sprintf("Positions %s", JSON))
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
