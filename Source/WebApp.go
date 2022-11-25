package MazeGameServer

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
)

type WebApp struct {
	Connection net.Conn
	MazeHost   MazeHost
}

func CreateWebApp(Connection net.Conn, Maze MazeHost) WebApp {
	return WebApp{
		Connection: Connection,
		MazeHost:   Maze,
	}
}

func (m WebApp) SendMessage(message string) error {
	_, err := m.Connection.Write([]byte(message))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to send message to WebApp: %s", err.Error()))
	}

	return nil
}

func (m WebApp) HandleMessages() {
	buffer, err := bufio.NewReader(m.Connection).ReadBytes('\n')
	if err != nil {
		// Client disconnected
		err := m.Disconnected()
		if err != nil {
			fmt.Printf("Failed to disconnect client: %s", err.Error())
		}
		return
	}

	var clientMessage = string(buffer[:len(buffer)-1])

	log.Println("Client message:", clientMessage)

	m.HandleMessages()
}

func (m WebApp) Disconnected() error {
	err := m.Connection.Close()
	if err != nil {
		return err
	}

	return nil
}

func (m WebApp) SetPlayerPosition(JSON string) {
	err := m.SendMessage(fmt.Sprintf("PlayerPosition %s", JSON))
	if err != nil {
		fmt.Printf("Error setting player position: %s", err.Error())
	}
}

func (m WebApp) SetMaze(JSON string) {
	err := m.SendMessage(fmt.Sprintf("Maze %s", JSON))
	if err != nil {
		fmt.Printf("Error setting maze: %s", err.Error())
	}
}
