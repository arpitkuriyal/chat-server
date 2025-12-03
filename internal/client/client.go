package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

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

	// 1) goroutine: read from server
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("read error from server:", err)
				return
			}
			fmt.Print("Server: ", string(buffer[:n]))
		}
	}()

	// 2) main goroutine: read from stdin and send to server
	stdin := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !stdin.Scan() {
			fmt.Println("stdin closed")
			return
		}
		text := stdin.Text() + "\n"

		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("write error to server:", err)
			return
		}
	}
}
