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
	clients   map[string]*Peer
	mu        sync.Mutex
}

type Peer struct {
	Conn     net.Conn
	Enc      *json.Encoder
	Dec      *json.Decoder
	Username string
	IsHost   bool
}

func NewSever() *Server {
	return &Server{
		broadcast: make(chan common.Message, 100), // why buffer channel
		clients:   make(map[string]*Peer),
		mu:        sync.Mutex{},
	}
}

func (s *Server) RunServer(addr string, ready chan<- bool) {
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
	listen, err := tls.Listen("tcp", addr, tlsConfig)
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
		for _, client := range s.clients {
			_ = client.Enc.Encode(msg)
		}
		s.mu.Unlock()
	}
}

// this function is call every time new client send message.
func (s *Server) handleClient(conn net.Conn) {
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)

	var client *Peer
	var username string

	// retry join loop
	for {
		var joinReq common.Message
		if err := dec.Decode(&joinReq); err != nil {
			conn.Close()
			return
		}

		if joinReq.Type != "join" {
			continue
		}

		username = joinReq.From

		s.mu.Lock()
		if _, exists := s.clients[username]; exists {
			_ = enc.Encode(common.Message{
				Type: "join-reject",
				Text: "username already taken, try another",
			})
			s.mu.Unlock()
			continue
		}

		// accept
		client = &Peer{
			Conn:     conn,
			Enc:      enc,
			Dec:      dec,
			Username: username,
			IsHost:   joinReq.IsHost,
		}

		s.clients[username] = client
		s.mu.Unlock()

		_ = enc.Encode(common.Message{
			Type: "join-accept",
		})
		break
	}

	if client.IsHost {
		s.broadcast <- common.Message{
			Type: "system",
			Text: fmt.Sprintf("%s (host) joined chat", username),
		}
	} else {
		s.broadcast <- common.Message{
			Type: "system",
			Text: fmt.Sprintf("%s joined chat", username),
		}
	}

	// clean up on disconnect
	defer func() {
		s.mu.Lock()
		delete(s.clients, username)
		s.mu.Unlock()
		conn.Close()
		s.broadcast <- common.Message{
			Type: "system",
			Text: fmt.Sprintf("%s left the chat", username),
		}
	}()

	for {
		var msg common.Message
		if err := dec.Decode(&msg); err != nil {
			return
		}

		if msg.Type == "chat" {
			s.broadcast <- msg
		}

		if msg.Text == "/exit" {
			return
		}
	}
}
