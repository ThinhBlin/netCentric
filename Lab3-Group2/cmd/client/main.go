package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

func runTCPOrders() {
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println("[client] TCP connect error:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	orders := []string{
		"ORDER|101|Noodles|2",
		"ORDER|102|Pho|1",
		"ORDER|103|Banh Mi|0", // Invalid order (quantity = 0)
	}

	for _, message := range orders {
		if _, err := fmt.Fprintln(conn, message); err != nil {
			fmt.Println("[client] TCP send error:", err)
			return
		}
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("[client] TCP read error:", err)
			return
		}
		fmt.Print("[client] ", strings.TrimSpace(response), "\n")
	}
}

func runUDPProbes() {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9001")
	if err != nil {
		fmt.Println("[client] UDP address error:", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("[client] UDP dial error:", err)
		return
	}
	defer conn.Close()

	probeIDs := []string{"probe1", "probe2", "probe3", "loss1", "drop"}
	for _, reqID := range probeIDs {
		sendProbeWithRetry(conn, reqID)
	}
}

func sendProbeWithRetry(conn *net.UDPConn, reqID string) {
	const maxAttempts = 2
	message := "PING|" + reqID
	expectedReply := "PONG|" + reqID + "|READY"

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		start := time.Now()

		if _, err := conn.Write([]byte(message)); err != nil {
			fmt.Println("[client] UDP send error:", err)
			return
		}

		if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			fmt.Println("[client] Deadline error:", err)
			return
		}

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Printf("[client] %s attempt %d timed out (1s)\n", reqID, attempt)
				continue
			}
			fmt.Println("[client] UDP network error:", err)
			return
		}

		reply := strings.TrimSpace(string(buffer[:n]))
		if reply == expectedReply {
			fmt.Printf("[client] %s (attempt %d, RTT %v)\n", reply, attempt, time.Since(start))
			return
		} else {
			fmt.Printf("[client] Wrong UDP reply ID: got %q, expected %q\n", reply, expectedReply)
			return
		}
	}
	fmt.Printf("[client] %s: NO RESPONSE\n", reqID)
}

func runHomeworkTests() {
	// Test 1: Concurrent TCP Submissions
	var wg sync.WaitGroup
	fmt.Println("\n[Test 1] Sending concurrent TCP submissions (IDs 201-205)...")
	for i := 201; i <= 205; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sendTCPOrder(fmt.Sprintf("ORDER|%d|Pho|1", id))
		}(i)
	}
	wg.Wait()
	time.Sleep(500 * time.Millisecond) // Brief pause to ensure server processes prints

	// Test 2: Duplicate TCP Order ID
	fmt.Println("\n[Test 2] Testing Duplicate TCP ID (resending ID 201)...")
	sendTCPOrder("ORDER|201|Pho|1")

	// Test 3: Known ID UDP Check
	fmt.Println("\n[Test 3] UDP Check - Known ID (201)")
	sendUDPCheck("CHECK|201")

	// Test 4: Unknown ID UDP Check
	fmt.Println("\n[Test 4] UDP Check - Unknown ID (999)")
	sendUDPCheck("CHECK|999")

	// HW Test 5: Malformed ID UDP Check
	fmt.Println("\n[Test 5] UDP Check - Malformed ID (CHECK|abc)")
	sendUDPCheck("CHECK|abc")
}

func sendTCPOrder(req string) {
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println("[TCP Client] Connect error:", err)
		return
	}
	defer conn.Close()

	fmt.Fprintln(conn, req)
	reader := bufio.NewScanner(conn)
	if reader.Scan() {
		fmt.Printf("[TCP Client] Sent: %s  -> Reply: %s\n", strings.TrimSpace(req), reader.Text())
	}
}

func sendUDPCheck(req string) {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9001")
	conn, _ := net.DialUDP("udp", nil, addr)
	defer conn.Close()

	conn.Write([]byte(req))
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("[UDP Client] Timeout or Error:", err)
		return
	}
	fmt.Printf("[UDP Client] Sent: %s -> Reply: %s\n", req, string(buf[:n]))
}

func main() {
	fmt.Println("=== INCLASS TESTS ===")
	fmt.Println("--- Running TCP Orders ---")
	runTCPOrders()

	fmt.Println("\n--- Running UDP Probes ---")
	runUDPProbes()

	fmt.Println("\n=== HOMEWORK TESTS ===")
	runHomeworkTests()
}