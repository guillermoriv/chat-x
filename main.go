package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

type Client struct {
	Conn net.Conn
	Name string
}

var (
	mutex     sync.Mutex
	clients   = make(map[net.Conn]Client)
	broadcast = make(chan string)
	names     = make(map[string]struct{})
)

func main() {
	port := ":9000"

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Println("❌ Failed to bind to port:", err)
		os.Exit(1)
	}

	defer listener.Close()

	log.Println("🚀 TCP server listening on", port)

	go handleBroadcast()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("⚠️ Error accepting connection:", err)
			continue
		}

		log.Println("✅ New client connected:", conn.RemoteAddr())
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer func() {
		log.Println("❌ Client disconnected:", conn.RemoteAddr())
		mutex.Lock()
		client := clients[conn]
		broadcast <- fmt.Sprintf("%s has disconnected from the chat.\n", client.Name)
		delete(clients, conn)
		delete(names, client.Name)
		mutex.Unlock()

		conn.Close()
	}()

	fmt.Fprintln(conn, "Welcome to Chat-X! Please enter your name: ")
	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Println("❌ Failed to read name from", conn.RemoteAddr())
		return
	}

	if name == "" || len(name) < 2 || len(name) > 16 {
		fmt.Fprintln(conn, "⚠️ Name must be at least 2 characters and less than 16 characters.")
		return
	}

	name = strings.TrimSpace(name)

	if _, exists := names[name]; exists {
		fmt.Fprintln(conn, "⚠️ Name is already taken. Try another one.")
		return
	}

	mutex.Lock()
	clients[conn] = Client{Conn: conn, Name: name}
	names[name] = struct{}{}
	mutex.Unlock()

	fmt.Fprintln(conn, "Welcome to Chat-X, let's have some fun!")
	fmt.Fprintln(conn, "To exit the chat, you can either use !quit or !exit, have fun.")

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		msg := scanner.Text()

		if msg == "!exit" || msg == "!quit" {
			log.Printf("👋 %s sent !quit, closing connection\n", conn.RemoteAddr())
			return
		}

		tagged := fmt.Sprintf("[%s]: %s\n", name, msg)
		broadcast <- tagged
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("📛 Read error:", err)
	}
}

func handleBroadcast() {
	for msg := range broadcast {
		mutex.Lock()
		for conn, client := range clients {
			if _, err := fmt.Fprintf(conn, "%s", msg); err != nil {
				log.Printf("❌ Failed to write to %s (%s): %v\n", client.Name, conn.RemoteAddr(), err)

				// Clean up dead client
				conn.Close()
				delete(clients, conn)
				delete(names, client.Name)

				// Tell others this client dropped
				go func(name string) {
					broadcast <- fmt.Sprintf("⚠️ %s has been disconnected unexpectedly.\n", name)
				}(client.Name)
			}
		}
		mutex.Unlock()
	}
}
