package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	connHost = "localhost"
	connPort = "25566"
	connType = "tcp"
)

func main() {
	fmt.Println("Starting " + connType + " server on " + connHost + ":" + connPort)
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		fmt.Println("Client connected.")

		fmt.Println("Client " + c.RemoteAddr().String() + " connected.")

		go handleConnection(c)
	}
}

func sendMessage(conn net.Conn, message string) (int, error) {
	n, err := conn.Write([]byte(message))
	if err != nil {
		return n, errors.New(fmt.Sprintf("Failed to send message to client: %s", err.Error()))
	}

	return n, err
}

func handleConnection(conn net.Conn) {
	buffer, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		fmt.Println("Client left.")
		conn.Close()
		return
	}

	log.Println("Client message:", string(buffer[:len(buffer)-1]))

	_, err = sendMessage(conn, "Hello from server!")
	if err != nil {
		fmt.Println(err.Error())
	}

	handleConnection(conn)
}
