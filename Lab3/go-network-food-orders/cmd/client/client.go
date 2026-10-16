package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func main() {
	// 1. Establish a TCP connection to the server
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// 2 & 3. Two valid orders and one invalid order on the same connection
	orders := []string{
		"ORD-101,Pho,2",        // Valid order 1
		"ORD-102,BanhMi,1",     // Valid order 2
		"ORD-103,Coffee,wrong", // Invalid order (bad quantity)
	}

	for _, order := range orders {
		fmt.Println("Sending:", order)
		fmt.Fprintln(conn, order) // Send complete line to server

		// Read the reply from server
		reply, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading reply:", err)
			return
		}
		fmt.Print("Server reply: ", reply)

		// 4. Short delay so you have time to run a 2nd client simultaneously in another terminal
		time.Sleep(2 * time.Second)
	}
}
