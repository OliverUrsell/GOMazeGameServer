package main

import (
	MazeGameServer "MazeGameServer/Source"
	"bufio"
	"fmt"
	"net"
	"os"
)

const (
	connHost = "localhost"
	connPort = "25566"
	connType = "tcp"
)

var CodeMazeHostMap map[string]MazeGameServer.MazeHost

func main() {
	fmt.Println("Starting " + connType + " server on " + connHost + ":" + connPort)
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	CodeMazeHostMap = make(map[string]MazeGameServer.MazeHost)

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
	buffer, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		fmt.Println("Client left.")
		conn.Close()
		return
	}

	var ClientMessage = string(buffer[:len(buffer)-1])

	// Check for (StartGame [Code] [MazeNodeJSON]) message from MazeHost
	if len(ClientMessage) > 15 && ClientMessage[:9] == "StartGame" {
		var Code = ClientMessage[10:14]
		var MazeJSON = ClientMessage[15:]

		fmt.Printf("New Maze Host has joined with code: %s\n", Code)
		fmt.Printf("JSON: %s\n", MazeJSON)

		_, ok := CodeMazeHostMap[Code]
		if ok {
			fmt.Println("Maze with code ", Code, " already exists!")
			return
		}

		CodeMazeHostMap[Code] = MazeGameServer.CreateHost(conn, Code, MazeJSON)
		return
	}

	// Check for the WebApp JoinGame message
	if len(ClientMessage) == 13 && ClientMessage[len(ClientMessage)-8:] == "JoinGame" {
		var Code = ClientMessage[:4]
		fmt.Printf("New Web App has joined with code: %s\n", Code)

		mh, ok := CodeMazeHostMap[Code]
		if !ok {
			// There is no host for that code
			_, err := conn.Write([]byte("NoMaze"))
			if err != nil {
				fmt.Printf("Failed to send message to WebApp: %s\n", err.Error())
			}
			conn.Close()
			return
		}

		mh.AddWebApp(MazeGameServer.CreateWebApp(conn, mh))

		return
	}

	fmt.Printf("Closing connection due to Unexpected message recieved on connect: %s\n", ClientMessage)
	conn.Close()
}
