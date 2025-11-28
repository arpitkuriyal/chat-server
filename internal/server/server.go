package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

var (
	broadcast = make(chan string)
	clients   = make(map[net.Conn]bool)
)

func main() {
	listen, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error in server main:", err)
		return
	}
	defer listen.Close()

	fmt.Println("server is listening on port 8080")
	go handleBroadcast()
	// accept the clients to communicate
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		clients[conn] = true
		fmt.Println("client connected:", conn.RemoteAddr())
		go handleClient(conn)
	}
}

func handleBroadcast() {
	for {
		msg := <-broadcast

		for conn := range clients {
			_, err := conn.Write([]byte(msg))
			if err != nil {
				conn.Close()
				delete(clients, conn)
				return
			}
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	// continuously read from client
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("read error:", err)
				return
			}
			fmt.Print("Client: ", string(buffer[:n]))
		}
	}()

	// read from server terminal (stdin) and send to client
	// it read line by line so we have to collect it in text variable and
	stdin := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Server: ")
		if !stdin.Scan() {
			fmt.Println("server stdin closed")
			return
		}
		text := stdin.Text() + "\n"
		broadcast <- text
	}
}
