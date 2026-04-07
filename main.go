package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	"golang.org/x/term"
	"os"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "7452"
	SERVER_TYPE = "tcp"
)

var (
	clients = make(map[string]*CLIENT)
	mu      sync.Mutex
)

type CLIENT struct {
	IP       string
	USERNAME string
	CONN     net.Conn
	MESSAGE  string
	ACTIVE   bool
	ANNONYM  bool
}


func parseMessage(raw string, conn net.Conn) (username, text string, annonym bool, quit bool) {
	raw = strings.TrimRight(raw, "\r\n")

	if !strings.HasPrefix(raw, "/") {
		return "", "", false, false
	}

	body := raw[1:]

	if body == "quit" {
		return "", "", false, true
	}

	parts := strings.SplitN(body, ":", 2)
	if len(parts) != 2 {
		conn.Write([]byte("Usage: /username:message  or  /annonym:message  or  /quit\n"))
		return "", "", false, false
	}

	username = strings.TrimSpace(parts[0])
	text = strings.TrimSpace(parts[1])

	if strings.ToLower(username) == "annonym" || username == "" {
		annonym = true
		username = "protected"
	}

	return username, text, annonym, false
}

func printRight(conn net.Conn, sender *CLIENT, displayName string) {
    
    width, _, _ := term.GetSize(int(os.Stdout.Fd()))
    if width < 20 {
        width = 80 
    }
    rightText := fmt.Sprintf("%s:[%s]", sender.MESSAGE, displayName)
    col := width - len(rightText) - 1 

    if col < 1 {
        col = 1
    }

    line := fmt.Sprintf("\033[1G\033[%dG%s\n", col, rightText)
    conn.Write([]byte(line))
}

func broadcast(senderIP string) {
	mu.Lock()
	defer mu.Unlock()

	sender := clients[senderIP]
	displayName := sender.USERNAME
	if sender.ANNONYM {
		displayName = "protected"
	}

	

	for ip, client := range clients {
		if ip != senderIP && client.ACTIVE {
			printRight(client.CONN, sender, displayName)
		}
	}
}

func handleClient(conn net.Conn) {
	buf := make([]byte, 1024)
	ip := conn.RemoteAddr().String()

	mu.Lock()
	clients[ip] = &CLIENT{IP: ip, ACTIVE: true, CONN: conn}
	mu.Unlock()

	conn.Write([]byte("Format: /username:message | /annonym:message | /quit\n"))

	for {
		n, err := conn.Read(buf)
		if err != nil {
			mu.Lock()
			clients[ip].ACTIVE = false
			mu.Unlock()
			conn.Close()
			return
		}

		raw := string(buf[:n])
		username, text, annonym, quit := parseMessage(raw, conn)

		if quit {
			mu.Lock()
			clients[ip].ACTIVE = false
			mu.Unlock()
			conn.Close()
			return
		}

		if text == "" {
			continue 
		}

		mu.Lock()
		clients[ip].USERNAME = username
		clients[ip].ANNONYM = annonym
		clients[ip].MESSAGE = text
		mu.Unlock()

		fmt.Printf("%s [%s] %s: %s\n",
			time.Now().Format("15:04"), ip, username, text)

		broadcast(ip)
	}
}

func main() {
	listener, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server started — listening on port " + SERVER_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}