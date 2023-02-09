package MazeGameServer

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"time"
)

type MazeHost struct {
	Connection                 net.Conn
	Reader                     *bufio.Reader
	Code                       string
	MazeJson                   string
	Positions                  string
	WebApps                    []*WebApp
	MonsterControllerConnected bool
}

func CreateHost(Connection net.Conn, Reader *bufio.Reader, Code, MazeJson string) *MazeHost {
	m := MazeHost{
		Connection:                 Connection,
		Reader:                     Reader,
		Code:                       Code,
		MazeJson:                   MazeJson,
		MonsterControllerConnected: false,
	}

	go m.HandleMessages()

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

	return errors.New(fmt.Sprintf("Failed after %s attempts", attempts))

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

func (m *MazeHost) HandleMessages() {
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

	//log.Println("Host ", m.Code, " message:", clientMessage)

	if len(clientMessage) > 5 && clientMessage[:4] == "Maze" {
		m.SetMaze(clientMessage[5:])
	}

	if len(clientMessage) > 10 && clientMessage[:9] == "Positions" {
		m.SetWebAppsPositions(clientMessage[9:])
	}

	m.HandleMessages()
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

func (m *MazeHost) AddWebApp(w *WebApp) {
	m.WebApps = append(m.WebApps, w)
}

func (m *MazeHost) SetWebAppsPositions(JSON string) {
	for i, app := range m.WebApps {
		err := app.SetPositions(JSON)
		if err != nil {
			fmt.Printf("Error setting positions: %s", err.Error())
			m.WebApps[i] = m.WebApps[len(m.WebApps)-1]
			m.WebApps = m.WebApps[:len(m.WebApps)-1]
		}
	}
}

func (m *MazeHost) SetMaze(JSON string) {
	m.MazeJson = JSON[5:]
	for i, app := range m.WebApps {
		err := app.SetMaze(JSON)
		if err != nil {
			fmt.Printf("Error setting maze: %s", err.Error())
			m.WebApps = append(m.WebApps[:i-1], m.WebApps[i:]...)
		}
	}
}

func (m *MazeHost) ChangeMonsterDirection(message string) error {
	err := m.SendMessage(message)
	if err != nil {
		return err
	}

	return nil
}
