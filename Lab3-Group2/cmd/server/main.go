package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	orderDB = make(map[int]string)
	dbMutex sync.Mutex
)

func main() {
	go startUDPServer()
	startTCPServer()
}

func startTCPServer() {
	listener, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println("[server] TCP listen error:", err)
		return
	}
	defer listener.Close()
	fmt.Println("[server] TCP listening on 127.0.0.1:9000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[server] TCP accept error:", err)
			continue
		}
		go handleTCPClient(conn)
	}
}

func handleTCPClient(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		reply := processOrderLine(line)
		conn.Write([]byte(reply))
	}
}

func processOrderLine(line string) string {
	parts := strings.Split(line, "|")
	if len(parts) != 4 || parts[0] != "ORDER" {
		return "ERR|INVALID_ORDER\n"
	}

	id, err1 := strconv.Atoi(parts[1])
	qty, err2 := strconv.Atoi(parts[3])
	food := strings.TrimSpace(parts[2])

	if err1 != nil || id <= 0 || err2 != nil || qty <= 0 || food == "" {
		return "ERR|INVALID_ORDER\n"
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()


	if _, exists := orderDB[id]; exists {
		fmt.Printf("[server] rejected duplicate order ID: %d\n", id)
		return "ERR|DUPLICATE_ID\n"
	}

	orderDB[id] = "RECEIVED"
	fmt.Printf("[server] order %d %s x%d\n", id, food, qty)
	return fmt.Sprintf("OK|%d|RECEIVED\n", id)
}

func startUDPServer() {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9001")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()
	fmt.Println("[server] UDP listening on 127.0.0.1:9001")

	buf := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg := strings.TrimSpace(string(buf[:n]))
		parts := strings.Split(msg, "|")

		// Homework: Handle CHECK|id
		if len(parts) == 2 && parts[0] == "CHECK" {
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				// Malformed ID
				conn.WriteToUDP([]byte("STATUS|MALFORMED_ID|NOT_FOUND"), clientAddr)
				continue
			}

			dbMutex.Lock()
			status, exists := orderDB[id]
			dbMutex.Unlock()

			if exists {
				reply := fmt.Sprintf("STATUS|%d|%s", id, status)
				conn.WriteToUDP([]byte(reply), clientAddr)
			} else {
				reply := fmt.Sprintf("STATUS|%d|NOT_FOUND", id)
				conn.WriteToUDP([]byte(reply), clientAddr)
			}
			continue
		}

		// Legacy Lab 03 PING handling
		if len(parts) == 2 && parts[0] == "PING" {
			reply := fmt.Sprintf("PONG|%s|READY", parts[1])
			conn.WriteToUDP([]byte(reply), clientAddr)
		}
	}
}