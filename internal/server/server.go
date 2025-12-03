package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"os"
)

var (
	broadcast = make(chan string)
	clients   = make(map[net.Conn]bool)
)

func main() {
	// it loads the publlic certificate and private key. It prove that i am the real server.
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		println("failed to load server key pair", err)
		return
	}

	// store the certificate so `tls.listen()` uses it during handshake
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	listen, err := tls.Listen("tcp", "localhost:8080", tlsConfig)
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
