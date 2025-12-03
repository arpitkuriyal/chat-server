package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
)

type Message struct {
	From string `json:"from"`
	Text string `json:"text"`
}

func main() {
	// load the CA cert so client trusts your server cert
	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		fmt.Println("failed to read ca cert")
	}

	// Create a certificate pool (trust store, empty bag) that will hold our trusted CA certificate. The client will use this pool to verify the server's TLS certificate.
	caPool := x509.NewCertPool()

	// Add our custom CA certificate (ca.crt) into caPool.
	if !caPool.AppendCertsFromPEM(caCert) {
		fmt.Println("failed to add CA cert to pool")
		return
	}

	// tls.Config is the “brain” of the TLS connection. It tells Go how to perform the TLS handshake and what security rules to use. Different for both client and server as they have different purpose to verify during handshake
	tlsConfig := &tls.Config{
		RootCAs:    caPool,      // tells the client which certificate authority to trust
		ServerName: "localhost", // must match server certifiate's SAN
	}
	conn, err := tls.Dial("tcp", "localhost:8080", tlsConfig)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	fmt.Println("connected to server")

	enc := json.NewEncoder(conn) // why
	dec := json.NewDecoder(conn) //why

	stdin := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter your username:")
	stdin.Scan()
	username := stdin.Text()

	// 1) goroutine: read from server
	go func() {
		for {
			var msg Message
			if err := dec.Decode(&msg); err != nil {
				fmt.Println("server disconnected", err)
				return
			}

			fmt.Printf("\n[%s] %s\n", msg.From, msg.Text)
		}
	}()

	// 2) main goroutine: read from stdin and send to server
	for {
		fmt.Print("You: ")
		if !stdin.Scan() {
			fmt.Println("stdin closed")
			return
		}

		msg := Message{
			From: username,
			Text: stdin.Text(),
		}

		if err := enc.Encode(msg); err != nil {
			fmt.Println("send error", err)
			return
		}
	}
}
