package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/arpitkuriyal/chat-server/internal/common"
)

type Server struct {
	broadcast chan common.Message
	clients   map[string]*Client
	mu        sync.Mutex
}

type Client struct {
	Conn     net.Conn
	Enc      *json.Encoder
	Dec      *json.Decoder
	Username string
	IsHost   bool
}

func NewSever() *Server {
	return &Server{
		broadcast: make(chan common.Message),
		clients:   make(map[string]*Client),
		mu:        sync.Mutex{},
	}
}

func (s *Server) RunServer(ready chan<- bool) {
	// it loads the public certificate and private key. It prove that i am the real server.
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
	ready <- true
	// when message come send to all client simaltaneously.
	go s.handleBroadcast()

	// accept the clients to communicate
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		go s.handleClient(conn)
	}
}

// broadcast message to all cilents
func (s *Server) handleBroadcast() {
	for msg := range s.broadcast {
		s.mu.Lock()
		for username, client := range s.clients {
			if err := client.Enc.Encode(msg); err != nil {
				client.Conn.Close()
				delete(s.clients, username)
				return
			}
			s.mu.Unlock()
		}
	}
}

// this function is call every time new client send message.
func (s *Server) handleClient(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	dec := json.NewDecoder(conn)

	for {
		var msg common.Message
		if err := dec.Decode(&msg); err != nil {
			// client disconnected unexpectedly
			s.broadcast <- common.Message{
				From: "system",
				Text: fmt.Sprintf("%s left the chat", msg.From),
			}
			return
		}

		// handle /exit command
		if msg.Text == "/exit" {
			s.broadcast <- common.Message{
				From: "system",
				Text: fmt.Sprintf("%s left the chat", msg.From),
			}
			return
		}

		s.broadcast <- msg
	}
}
