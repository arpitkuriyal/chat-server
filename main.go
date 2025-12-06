package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/arpitkuriyal/chat-server/internal/client"
	"github.com/arpitkuriyal/chat-server/internal/server"
)

func main() {
	host := flag.Bool("host", false, "run in host mode (start server + client)")
	addr := flag.String("addr", "localhost:8080", "server address to listen/connect")
	flag.Parse()

	if *host {
		fmt.Println("Starting server on", *addr)
		go server.RunServer()

		// small delay so server starts listening
		time.Sleep(300 * time.Millisecond)

		fmt.Println("Starting host client...")
		client.Run(*addr, true)
	} else {
		fmt.Println("Starting client, connecting to", *addr)
		client.Run(*addr, false)
	}
}
