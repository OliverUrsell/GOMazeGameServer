package MazeGameServer

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
)

type MazeHost struct {
	Connection     net.Conn
	Code           string
	MazeJson       string
	PlayerPosition string
	WebApps        []WebApp
}

func CreateHost(Connection net.Conn, Code, MazeJson string) MazeHost {
	m := MazeHost{
		Connection: Connection,
		Code:       Code,
		MazeJson:   MazeJson,
	}

	go m.HandleMessages()

	return m
}

func (m MazeHost) SendMessage(message string) error {
	_, err := m.Connection.Write([]byte(message))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to send message to MazeHost: %s", err.Error()))
	}

	return nil
}

func (m MazeHost) HandleMessages() {
	buffer, err := bufio.NewReader(m.Connection).ReadBytes('\n')
	if err != nil {
		// Client disconnected
		err := m.Disconnected()
		if err != nil {
			fmt.Printf("Failed to disconnect client: %s\n", err.Error())
		}

		fmt.Printf("Host with code: %s has disconnected\n", m.Code)
		return
	}

	var clientMessage = string(buffer[:len(buffer)-1])

	log.Println("Host ", m.Code, " message:", clientMessage)

	if len(clientMessage) > 5 && clientMessage[:4] == "Maze" {
		m.SetMaze(clientMessage[5:])
	}

	if len(clientMessage) > 15 && clientMessage[:14] == "PlayerPosition" {
		m.SetWebAppsPlayerPosition(clientMessage[15:])
	}

	m.HandleMessages()
}

func (m MazeHost) Disconnected() error {
	err := m.Connection.Close()
	if err != nil {
		return err
	}

	return nil
}

func (m MazeHost) AddWebApp(w WebApp) {
	m.WebApps = append(m.WebApps, w)
}

func (m MazeHost) SetWebAppsPlayerPosition(JSON string) {
	for _, app := range m.WebApps {
		app.SetPlayerPosition(JSON)
	}
}

func (m MazeHost) SetMaze(JSON string) {
	m.MazeJson = JSON[5:]
	for _, app := range m.WebApps {
		app.SetMaze(JSON)
	}
}
