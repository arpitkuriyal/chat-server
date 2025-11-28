package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	listen, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error in server main:", err)
		return
	}
	defer listen.Close()

	fmt.Println("server is listening on port 8080")

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		fmt.Println("client connected:", conn.RemoteAddr())
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	// 1) goroutine: continuously read from client
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

	// 2) main part: read from server terminal (stdin) and send to client
	stdin := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Server: ")
		if !stdin.Scan() {
			fmt.Println("server stdin closed")
			return
		}
		text := stdin.Text() + "\n"
		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("write error:", err)
			return
		}
	}
}
