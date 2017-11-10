package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"code.cloudfoundry.org/rfc5424"
)

func main() {
	l, err := net.Listen("tcp4", fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		go func() {
			defer conn.Close()
			handleConnection(conn)
		}()
	}
}

func handleConnection(conn net.Conn) {
	msg := rfc5424.Message{}
	for {
		if _, err := msg.ReadFrom(conn); err != nil {
			return
		}

		fmt.Println(string(msg.Message))
	}
}
