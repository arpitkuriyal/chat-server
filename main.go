package main

import (
	"flag"
	"fmt"

	"github.com/arpitkuriyal/chat-server/internal/client"
	"github.com/arpitkuriyal/chat-server/internal/server"
)

func main() {
	host := flag.Bool("host", false, "run as server")
	flag.Parse()

	if *host {
		fmt.Println("Starting server...")
		server.RunServer()
	} else {
		fmt.Println("Starting client...")
		client.RunClient()
	}
}
