package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("udp", "127.0.0.1:9001")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 512)
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("req-%d", i)

		start := time.Now()
		if _, err := conn.Write([]byte("PING|" + id)); err != nil {
			fmt.Printf("%s: send error: %v\n", id, err)
			continue
		}

		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := conn.Read(buf)
		rtt := time.Since(start)

		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				fmt.Printf("%s: no response (timeout)\n", id)
			} else {
				fmt.Printf("%s: error: %v\n", id, err)
			}
			continue
		}

		want := fmt.Sprintf("PONG|%s|READY", id)
		if got := string(buf[:n]); got != want {
			fmt.Printf("%s: invalid reply %q\n", id, got)
			continue
		}
		fmt.Printf("%s: OK, RTT = %v\n", id, rtt)
	}
}
