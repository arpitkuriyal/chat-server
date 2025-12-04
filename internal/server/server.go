package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Message struct {
	From string `json:"from"`
	Text string `json:"text"`
}

var (
	broadcast = make(chan Message)
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

	// when message come send to all client simaltaneously.
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

// broadcast message to all cilents
func handleBroadcast() {
	for {
		msg := <-broadcast

		for conn := range clients {
			enc := json.NewEncoder(conn)
			if err := enc.Encode(msg); err != nil {
				conn.Close()
				delete(clients, conn)
				return
			}
		}
	}
}

// this function is call every time new client send message.
func handleClient(conn net.Conn) {
	defer conn.Close()
	dec := json.NewDecoder(conn)

	// continuously read from client
	go func() {
		for {
			var msg Message

			// make the JSON data in go struct and store it in msg
			err := dec.Decode(&msg)
			if err != nil {
				fmt.Println("Client disconnected:", err)
				conn.Close()
				delete(clients, conn)
				return
			}

			// first server sees the messagae after that we broadcast it to all client thats why `broadcast <- msg` is after this.
			fmt.Printf("Client [%s]: %s\n: ", msg.From, msg.Text)

			broadcast <- msg
		}
	}()

	// read from server terminal (stdin) and send to client
	stdin := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Server: ")
		if !stdin.Scan() {
			fmt.Println("server stdin closed")
			return
		}
		msg := Message{
			From: "SERVER",
			Text: stdin.Text(),
		}

		// broadcast to all the client
		broadcast <- msg
	}
}
