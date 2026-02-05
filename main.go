package main

import (
	"flag"
	"fmt"

	"github.com/arpitkuriyal/chat-server/internal/client"
	"github.com/arpitkuriyal/chat-server/internal/server"
)

func main() {
	host := flag.Bool("host", false, "run in host mode (start server + client)")
	addr := flag.String("addr", "localhost:8080", "server address to listen/connect")
	flag.Parse()

	if *host {
		fmt.Println("Starting server on", *addr)
		ready := make(chan bool)
		srv := server.NewSever()
		go srv.RunServer(*addr, ready)
		<-ready // waiting until server is ready

		fmt.Println("Starting host client...")
		client.Run(*addr, true)
	} else {
		fmt.Println("Starting client, connecting to", *addr)
		client.Run(*addr, false)
	}
}
