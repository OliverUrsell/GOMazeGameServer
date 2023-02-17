package main

import (
	MazeGameServer "MazeGameServer/Source"
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

const (
	connHost = "0.0.0.0"
	connPort = "25566"
	connType = "tcp4"
)

var CodeMazeHostMap map[string]*MazeGameServer.MazeHost

func main() {
	// Start a server for the Unreal Engine mazes
	go startSocketServer()

	// Start a server for web sockets
	startWebSocketServer()
}

func startWebSocketServer() {
	fmt.Println("Starting WebSocket server on localhost 25567")

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		wsocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Websocket Connected!")
		listen(wsocket)
	})
	err := http.ListenAndServe(":25567", nil)
	if err != nil {
		log.Fatal("Failed to start server: " + err.Error())
	}
}

func listen(conn *websocket.Conn) {
	for {
		// read a message
		messageType, messageContent, _ := conn.ReadMessage()

		var ClientMessage = string(messageContent)

		// print out that message
		fmt.Println(ClientMessage)

		// Check for the WebApp JoinGame message
		if len(ClientMessage) == 14 && ClientMessage[len(ClientMessage)-9:] == "JoinGame\n" {
			var Code = ClientMessage[:4]
			fmt.Printf("New Web App has joined with code: %s\n", Code)

			mh, ok := CodeMazeHostMap[Code]
			if !ok {
				// There is no host for that code
				fmt.Println("There was no host for that code")
				if err := conn.WriteMessage(messageType, []byte("NoMaze\n")); err != nil {
					fmt.Printf("Failed to send message to WebApp: %s\n", err.Error())
				}
				return
			}

			webapp, err := MazeGameServer.CreateWebApp(conn, messageType, mh)
			if err != nil {
				fmt.Printf("Error creating a web app: %s\n", err.Error())
			}

			mh.AddWebApp(webapp)
		} else if len(ClientMessage) > len("1234 MonsterDirection") && strings.HasPrefix(ClientMessage[5:], "MonsterDirection") {
			// We got a monster direction message
			code := ClientMessage[:4]
			maze, ok := CodeMazeHostMap[code]
			if !ok {
				fmt.Println("No maze exists with code: " + code)
			}

			err := maze.ChangeMonsterDirection(ClientMessage)
			if err != nil {
				fmt.Println("Failed to set monster direction: " + err.Error())
			}
		}
	}
}

func startSocketServer() {
	fmt.Println("Starting " + connType + " server on " + connHost + ":" + connPort)
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	CodeMazeHostMap = make(map[string]*MazeGameServer.MazeHost)

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}

		fmt.Println("Client " + c.RemoteAddr().String() + " connected.")

		handleConnection(c)
	}
}

func handleConnection(conn net.Conn) {
	var reader = bufio.NewReader(conn)
	buffer, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Println("Client left.")
		conn.Close()
		return
	}

	var ClientMessage = string(buffer[:len(buffer)-1])

	// Check for (StartGame [Code] [MazeNodeJSON]) message from MazeHost
	if len(ClientMessage) == 14 && ClientMessage[:9] == "StartGame" {
		var Code = ClientMessage[10:14]

		fmt.Printf("New Maze Host has joined with code: %s\n", Code)

		_, ok := CodeMazeHostMap[Code]
		if ok {
			fmt.Println("Maze with code ", Code, " already exists!")
			conn.Close()
			return
		}

		CodeMazeHostMap[Code] = MazeGameServer.CreateHost(conn, reader, Code)
		return
	}

	fmt.Printf("Closing connection due to Unexpected message recieved on connect: %s\n", ClientMessage)
	conn.Close()
}
