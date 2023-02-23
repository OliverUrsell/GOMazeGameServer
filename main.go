package main

import MazeGameServer "MazeGameServer/Source"

func main() {
	// Start a server for the Unreal Engine mazes
	go MazeGameServer.StartSocketServer()

	// Start a server for web sockets
	MazeGameServer.StartWebSocketServer()
}
