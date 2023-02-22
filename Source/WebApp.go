package MazeGameServer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
)

type WebApp struct {
	Connection  *websocket.Conn
	MessageType int
	MazeHost    *MazeHost
	IsGuide     bool // If not a guide they're a monster controller
}

type IWebApp interface {
	SendMessage(message string) error
	SetPositions(JSON string) error
	Disconnected() error
	SetMaze(JSON string) error
	IsAGuide() bool
}

//go:generate moq -out IWebApp_mock.go . IWebApp

func CreateWebApp(Connection *websocket.Conn, MessageType int, Maze *MazeHost) (*WebApp, error) {
	var out = &WebApp{
		Connection:  Connection,
		MessageType: MessageType,
		MazeHost:    Maze,
		IsGuide:     Maze.MonsterControllerConnected,
	}

	if !out.IsGuide {
		Maze.MonsterControllerConnected = true
	}

	err := out.SetMaze(Maze.MazeJson)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (m WebApp) SendMessage(message string) error {
	//fmt.Printf("Sent Message to client: %s\n", message)
	if err := m.Connection.WriteMessage(m.MessageType, []byte(message+"\n")); err != nil {
		return errors.New(fmt.Sprintf("Failed to send message to WebApp: %s", err.Error()))
	}

	return nil
}

//func (m WebApp) HandleMessages(clientMessage string) {
//
//	log.Println("Client message: ", clientMessage)
//
//	if len(clientMessage) > 17 && clientMessage[:16] == "MonsterDirection" {
//
//		err := m.MazeHost.SendMessage(clientMessage)
//		if err != nil {
//			return
//		}
//	}
//}

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

	// If no maze has been set yet don't send a message
	if JSON == "nomaze" {
		return nil
	}

	// Add whether this player is a guide or monster to the json
	var mazeJSON map[string]interface{}

	err := json.Unmarshal([]byte(JSON), &mazeJSON)
	if err != nil {
		return err
	}

	if m.IsGuide {
		mazeJSON["player_type"] = "guide"
	} else {
		mazeJSON["player_type"] = "monster"
	}

	jsonBytes, err := json.Marshal(mazeJSON)
	if err != nil {
		return err
	}

	err = m.SendMessage("MAZE " + string(jsonBytes))
	if err != nil {
		return err
	}

	return nil
}

func (m WebApp) IsAGuide() bool {
	return m.IsGuide
}
