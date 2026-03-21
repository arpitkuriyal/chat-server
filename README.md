# Chat Server | Go

A secure, concurrent, terminal-based chat server built in Go using TCP and TLS.

---

## Features

- TLS-based secure communication  
  Ensures encrypted communication between client and server using a trusted certificate authority.

- Concurrent client handling using goroutines  
  Each client connection runs in its own goroutine, allowing multiple users to chat simultaneously.

- Channel-based message broadcasting  
  Messages are sent through a central broadcast channel and distributed to all connected clients.

- JSON-based client–server protocol  
  All communication is structured using JSON for simplicity and consistency.

- Terminal UI client using tview  
  Provides an interactive terminal interface with a chat window, input field, and active users panel.

- Unique username enforcement  
  The server ensures that no two clients can join with the same username.

- Active user list synchronization  
  The server maintains and broadcasts the list of connected users to all clients in real time.

- Basic command support:
  - `/nick <name>`  
    Changes your current username to `<name>` if it is not already taken.

  - `/whoami`  
    Displays your current username and whether you are the host.

  - `/exit`  
    Disconnects you from the server and closes the client session.

## Tech Stack

* Go (Golang)
* TCP Networking
* TLS / X.509 Certificates
* Goroutines & Channels
* JSON
* Terminal UI

---

## Usage

Run server:

```
go run main.go -host
```

Run client:

```
go run main.go
```

---
