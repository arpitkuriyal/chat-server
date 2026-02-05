package client

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/arpitkuriyal/chat-server/internal/common"
)

type Client struct {
	Conn     net.Conn
	Enc      *json.Encoder
	Dec      *json.Decoder
	Username string
	IsHost   bool
}

// Run connects to the server, builds a Client, then hands off to the TUI.
func Run(addr string, isHost bool) {
	client, err := NewClient(addr, isHost)
	if err != nil {
		fmt.Println("client error:", err)
		return
	}
	defer client.Conn.Close()

	if err := StartTUI(client); err != nil {
		fmt.Println("tui error:", err)
	}
}

func NewClient(addr string, isHost bool) (*Client, error) {
	// load the CA cert so client trusts your server cert
	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("failed to read ca cert: %w", err)
	}

	// Create a certificate pool (trust store, empty bag) that will hold our trusted CA certificate. The client will use this pool to verify the server's TLS certificate.
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA cert to pool")
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid addr: %w", err)
	}
	// tls.Config is the “brain” of the TLS connection. It tells Go how to perform the TLS handshake and what security rules to use. Different for both client and server as they have different purpose to verify during handshake
	tlsConfig := &tls.Config{
		RootCAs:    caPool, // tells the client which certificate authority to trust
		ServerName: host,   // must match server certifiate's SAN
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting: %w", err)
	}

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	// ask username in plain terminal (before TUI takes over)
	stdin := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter your username: ")
	stdin.Scan()
	username := strings.TrimSpace(stdin.Text())
	if username == "" {
		username = "anon"
	}
	if isHost {
		username = username + " (host)"
	}

	// do it for the first message "X joined the chat"
	handshake := common.Message{
		From:   username,
		Text:   "joined the chat",
		IsHost: isHost,
	}

	if err := enc.Encode(handshake); err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Enc:      enc,
		Dec:      dec,
		Username: username,
		IsHost:   isHost,
	}, nil
}

// helper so TUI code can send messages easily
func (c *Client) SendMessage(text string) error {
	msg := common.Message{
		From: c.Username,
		Text: text,
	}
	return c.Enc.Encode(msg)
}
