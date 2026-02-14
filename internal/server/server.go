package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"strings"
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
	Send     chan common.Message
}

func NewSever() *Server {
	return &Server{
		broadcast: make(chan common.Message, 100),
		clients:   make(map[string]*Peer),
	}
}

func (s *Server) RunServer(addr string, ready chan<- bool) {
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		fmt.Println("failed to load server key pair:", err)
		return
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listen, err := tls.Listen("tcp", addr, tlsConfig)
	if err != nil {
		fmt.Println("Error in server main:", err)
		return
	}
	defer listen.Close()

	fmt.Println("server is listening on", addr)
	ready <- true

	go s.handleBroadcast()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			return
		}
		go s.handleClient(conn)
	}
}

func (s *Server) handleBroadcast() {
	for msg := range s.broadcast {
		s.mu.Lock()
		for _, client := range s.clients {
			select {
			case client.Send <- msg:
			default:
			}
		}
		s.mu.Unlock()
	}
}

func (p *Peer) writeLoop() {
	for msg := range p.Send {
		_ = p.Enc.Encode(msg)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)

	var client *Peer
	var username string

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
				Text: "username already taken",
			})
			s.mu.Unlock()
			continue
		}

		client = &Peer{
			Conn:     conn,
			Enc:      enc,
			Dec:      dec,
			Username: username,
			IsHost:   joinReq.IsHost,
			Send:     make(chan common.Message, 50),
		}

		s.clients[username] = client
		s.mu.Unlock()

		go client.writeLoop()

		client.Send <- common.Message{Type: "join-accept"}
		s.sendUserList()
		break
	}

	s.broadcast <- common.Message{
		Type: "system",
		From: "system",
		Text: fmt.Sprintf("%s joined chat", username),
	}

	defer func() {
		s.mu.Lock()
		delete(s.clients, username)
		s.mu.Unlock()

		close(client.Send)
		conn.Close()

		s.sendUserList()
		s.broadcast <- common.Message{
			Type: "system",
			From: "system",
			Text: fmt.Sprintf("%s left the chat", username),
		}
	}()

	for {
		var msg common.Message
		if err := dec.Decode(&msg); err != nil {
			return
		}

		if strings.HasPrefix(msg.Text, "/") {
			s.handleCommand(client, &username, msg.Text)
			continue
		}

		if msg.Type == "chat" {
			s.broadcast <- msg
		}
	}
}

func (s *Server) handleCommand(p *Peer, username *string, text string) {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {

	case "/exit":
		p.Send <- common.Message{
			Type: "system",
			Text: "Goodbye",
		}
		p.Conn.Close()

	case "/nick":
		if len(parts) != 2 {
			p.Send <- common.Message{
				Type: "system",
				Text: "usage: /nick <newname>",
			}
			return
		}

		newName := parts[1]

		s.mu.Lock()
		if _, exists := s.clients[newName]; exists {
			s.mu.Unlock()
			p.Send <- common.Message{
				Type: "system",
				Text: "username already taken",
			}
			return
		}

		delete(s.clients, *username)
		s.clients[newName] = p
		oldName := *username
		*username = newName
		p.Username = newName
		s.mu.Unlock()

		s.sendUserList()
		s.broadcast <- common.Message{
			Type: "system",
			From: "system",
			Text: fmt.Sprintf("%s is now known as %s", oldName, newName),
		}

	default:
		p.Send <- common.Message{
			Type: "system",
			Text: "unknown command",
		}
	}
}

func (s *Server) sendUserList() {
	s.mu.Lock()
	defer s.mu.Unlock()

	users := make([]string, 0, len(s.clients))
	for name := range s.clients {
		users = append(users, name)
	}

	msg := common.Message{
		Type:  "user-list",
		Users: users,
	}

	for _, client := range s.clients {
		select {
		case client.Send <- msg:
		default:
		}
	}
}
