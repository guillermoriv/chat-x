# Chat-X 🗨️

**Chat-X** is a terminal-based TCP chat server written in Go. It allows multiple users to connect over the network, choose a name, and chat with each other in real time.

Built for simplicity and learning, this server uses raw TCP sockets, no web tech, and is fully cross-platform (Linux, macOS, Windows).

---

## 🚀 Features

- Accepts multiple clients over TCP
- Broadcasts messages to all connected users
- Each user must register a unique name (2–16 characters)
- Gracefully handles disconnections and command-based exits
- Server logs and client messages are clearly separated
- Includes basic error handling for broken connections

---

## 📦 Requirements

- Go 1.18+
- A terminal
- Clients can use:
  - `telnet`
  - `nc` (netcat)
  - a custom TCP client

---

## 🛠 How to Run

### Start the Server

```bash
go run main.go

```
