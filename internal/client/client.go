package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
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
