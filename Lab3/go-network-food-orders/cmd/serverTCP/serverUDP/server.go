package main

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// Rule: a request is valid only if it is exactly "PING|<id>" where
// <id> is 1-32 chars of letters, digits, '_' or '-'.
// Anything else is silently ignored (no reply).
var idRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,32}$`)

func parse(msg string) (string, bool) {
	parts := strings.Split(strings.TrimSpace(msg), "|")
	if len(parts) != 2 || parts[0] != "PING" || !idRe.MatchString(parts[1]) {
		return "", false
	}
	return parts[1], true
}

func main() {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9001")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("UDP server listening on", addr)

	buf := make([]byte, 512) // small datagrams only
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		id, ok := parse(string(buf[:n]))
		if !ok {
			fmt.Printf("ignored malformed from %v: %q\n", remote, buf[:n])
			continue
		}
		reply := fmt.Sprintf("PONG|%s|READY", id)
		conn.WriteToUDP([]byte(reply), remote)
		fmt.Printf("replied to %v: %s\n", remote, reply)
	}
}
