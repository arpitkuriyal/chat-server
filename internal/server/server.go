package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"

	"github.com/arpitkuriyal/chat-server/internal/common"
)

var (
	broadcast = make(chan common.Message)
	clients   = make(map[net.Conn]bool)
)

func RunServer() {
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

	// read from client
	for {
		var msg common.Message

		// make the JSON data in go struct and store it in msg
		err := dec.Decode(&msg)
		if err != nil {
			fmt.Println("Client disconnected:", err)
			conn.Close()
			delete(clients, conn)
			continue
		}

		broadcast <- msg
	}
}
