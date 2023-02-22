package MazeGameServer

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type MazeHost struct {
	Connection                 net.Conn
	Reader                     *bufio.Reader
	Code                       string
	MazeJson                   string
	Positions                  string
	WebApps                    []IWebApp
	MonsterControllerConnected bool
}

func CreateHost(Connection net.Conn, Reader *bufio.Reader, Code string) *MazeHost {
	m := MazeHost{
		Connection:                 Connection,
		Reader:                     Reader,
		Code:                       Code,
		MazeJson:                   "nomaze",
		MonsterControllerConnected: false,
	}

	go m.HandleMessages(true)

	return &m
}

func (m *MazeHost) _SendMessageLoop(message []byte) error {
	attempts := 0
	for attempts <= 10 {
		n, err := m.Connection.Write(message)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to send message to MazeHost: %s", err.Error()))
		}

		if n == len(message) {
			return nil
		}

		attempts += 1
		sleepTime, err := time.ParseDuration("10ms")
		if err != nil {
			return err
		}
		time.Sleep(sleepTime)
	}

	return errors.New(fmt.Sprintf("Failed after %d attempts", attempts))

}

func (m *MazeHost) SendMessage(message string) error {
	//fmt.Println("Sending message to host with code " + m.Code + ": " + message)
	// Need a message queue, if something fails to send (n is nil or something?) don't dequeue, keep sending, otherwise do dequeue
	go func() {
		err := m._SendMessageLoop([]byte(message))
		if err != nil {
			fmt.Printf("Failed to send message: " + err.Error())
		}
	}()

	//TODO: Fix that we have to return an error here
	return nil
}

func (m *MazeHost) HandleMessages(loop bool) {
	buffer, err := m.Reader.ReadBytes('\n')
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

	//log.Println("Host", m.Code, "message:", clientMessage)

	if len(clientMessage) > 10 && clientMessage[5:9] == "Maze" {
		//Code := clientMessage[:4]
		log.Println("Host ", m.Code, " message:", clientMessage)
		mazeJSON := clientMessage[10:]
		m.SetMaze(mazeJSON)
	}

	if len(clientMessage) > 13 && clientMessage[5:14] == "Positions" {
		m.SetWebAppsPositions(clientMessage[15:])
	}

	if loop {
		m.HandleMessages(loop)
	}
}

func (m *MazeHost) Disconnected() error {
	// TODO: Is this actually being called
	// TODO: Need to destroy this instance and remove it from the CodeMazeHostMap in main.go, possibly with an event paradigm
	err := m.Connection.Close()
	if err != nil {
		return err
	}

	return nil
}

func (m *MazeHost) AddWebApp(w IWebApp) {
	m.WebApps = append(m.WebApps, w)
}

func (m *MazeHost) SetWebAppsPositions(JSON string) {
	for i, app := range m.WebApps {
		err := app.SetPositions(JSON)
		if err != nil {
			fmt.Printf("Error setting positions: %s", err.Error())
			err = m.WebApps[i].Disconnected()
			if err != nil {
				fmt.Printf("Error disconnecting webapp: %s", err.Error())
			}
			m.WebApps[i] = m.WebApps[len(m.WebApps)-1]
			m.WebApps = m.WebApps[:len(m.WebApps)-1]
		}
	}
}

func (m *MazeHost) SetMaze(JSON string) {
	m.MazeJson = JSON
	for i, app := range m.WebApps {
		err := app.SetMaze(JSON)
		if err != nil {
			fmt.Printf("Error setting maze: %s", err.Error())
			err = m.WebApps[i].Disconnected()
			if err != nil {
				fmt.Printf("Error disconnecting webapp: %s", err.Error())
			}
			m.WebApps = append(m.WebApps[:i-1], m.WebApps[i:]...)
		}
	}

	// Wait a short period of time, so the webapps can pick up the maze message without any position messages in the buffer
	time.Sleep(100 * time.Millisecond)
}

func (m *MazeHost) ChangeMonsterDirection(message string) error {
	err := m.SendMessage(message)
	if err != nil {
		return err
	}

	return nil
}
