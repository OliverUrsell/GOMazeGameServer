package MazeGameServer

import (
	"bufio"
	"github.com/gorilla/websocket"
	"net"
	"reflect"
	"strings"
	"testing"
)

func CreateTestSocketServer() net.Conn {
	server, _ := net.Pipe()
	return server
}

func CreateTestWebSocketServer() *websocket.Conn {
	server := websocket.Conn{}
	return &server
}

func TestCreateHost(t *testing.T) {
	server := CreateTestSocketServer()
	defer server.Close()

	reader := bufio.NewReader(strings.NewReader("helloworld\n"))
	host := CreateHost(server, reader, "CODE")

	expectedHost := &MazeHost{
		Connection:                 server,
		Reader:                     reader,
		Code:                       "CODE",
		MazeJson:                   "nomaze",
		MonsterControllerConnected: false,
	}

	if !reflect.DeepEqual(host, expectedHost) {
		t.Errorf("Expected host: %v, got host: %v", *expectedHost, *host)
	}
}

func TestHandleMazeMessages(t *testing.T) {
	server := CreateTestSocketServer()
	defer server.Close()

	reader := bufio.NewReader(strings.NewReader("CODE Maze testjson\n"))
	host := CreateHost(server, reader, "CODE")
	host.HandleMessages(false)

	if host.MazeJson != "testjson" {
		t.Errorf("Expected json: testjson, got json: %s", host.MazeJson)
	}
}

func TestAddWebApp(t *testing.T) {
	hostserver := CreateTestSocketServer()
	defer hostserver.Close()

	webserver := CreateTestWebSocketServer()

	reader := bufio.NewReader(strings.NewReader("helloworld\n"))
	host := CreateHost(hostserver, reader, "CODE")

	webapp, err := CreateWebApp(webserver, 0, host)
	if err != nil {
		t.Errorf("Error creating webapp: %s", err.Error())
	}

	host.AddWebApp(webapp)

	if len(host.WebApps) != 1 {
		t.Errorf("Expected 1 webapp attached to host, got: %d", len(host.WebApps))
	}

	if !reflect.DeepEqual(host.WebApps[0], webapp) {
		t.Errorf("Expected webapp: %v, got webapp: %v", host.WebApps[0], webapp)
	}

	if host.WebApps[0].IsAGuide() != false {
		t.Errorf("Expected first webapp to be a monster, instead it was a guide")
	}

	webapp2, err := CreateWebApp(webserver, 0, host)
	if err != nil {
		t.Errorf("Error creating webapp: %s", err.Error())
	}

	host.AddWebApp(webapp2)

	if len(host.WebApps) != 2 {
		t.Errorf("Expected 2 webapps attached to host, got: %d", len(host.WebApps))
	}

	if !reflect.DeepEqual(host.WebApps[1], webapp2) {
		t.Errorf("Expected webapp: %v, got webapp: %v", host.WebApps[0], webapp2)
	}

	if host.WebApps[1].IsAGuide() != true {
		t.Errorf("Expected second webapp to be a guide, instead it was a monster")
	}

}

func TestHandlePositionsMessage(t *testing.T) {
	server := CreateTestSocketServer()
	defer server.Close()

	reader := bufio.NewReader(strings.NewReader("CODE Positions {\"hello\":\"world\"}\n"))
	host := CreateHost(server, reader, "CODE")

	mockWebApp := &IWebAppMock{
		SetPositionsFunc: func(JSON string) error {
			expectedJSON := "{\"hello\":\"world\"}"
			if expectedJSON != JSON {
				t.Errorf("Expected SetPositions to be called with %s, instead got %s", expectedJSON, JSON)
			}
			return nil
		},
	}

	host.AddWebApp(mockWebApp)

	host.HandleMessages(false)

	if len(mockWebApp.calls.SetPositions) != 1 {
		t.Errorf("webapp send message should be called once, instead it was called %d times.", len(mockWebApp.calls.SetPositions))
	}
}
