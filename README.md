# Chat Server | Go

A secure, concurrent, terminal-based chat server built in Go using TCP and TLS.

---

## Features

* Secure communication using TLS
* Multiple concurrent clients using goroutines
* Channel-based message broadcasting
* JSON-based client–server protocol
* Terminal-based TUI client
* Server-enforced unique usernames
* Host and client execution modes

---

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
