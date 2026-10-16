package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func handleClient(conn net.Conn) {
	// 4. Close connection and log client disconnect when finished
	defer func() {
		conn.Close()
		fmt.Printf("[DISCONNECT] Client %s disconnected\n", conn.RemoteAddr())
	}()

	fmt.Printf("[CONNECT] Client %s connected\n", conn.RemoteAddr())

	// 3. Read one complete line per order request
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Expecting format: OrderID,Item,Quantity (e.g., "ORD-101,Pho,2")
		fields := strings.Split(line, ",")
		if len(fields) != 3 || fields[0] == "" || fields[1] == "" {
			fmt.Fprintln(conn, "ERROR: Invalid format. Expected OrderID,Item,Quantity")
			continue
		}

		qty, err := strconv.Atoi(fields[2])
		if err != nil || qty <= 0 {
			fmt.Fprintln(conn, "ERROR: Quantity must be a positive number")
			continue
		}

		// 4. Log received order
		fmt.Printf("[ORDER] From %s -> ID: %s | Item: %s | Qty: %d\n",
			conn.RemoteAddr(), fields[0], fields[1], qty)

		// 3. Respond exactly once per request
		fmt.Fprintf(conn, "OK: Order %s (%d x %s) received\n", fields[0], qty, fields[1])
	}
}

func main() {
	// 1. Listen for TCP connections on 127.0.0.1:9000
	listener, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("TCP Order Server listening on 127.0.0.1:9000...")

	// 1 & 2. Accept multiple clients continuously, each in a separate goroutine
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go handleClient(conn)
	}
}
