package main

import (
	"fmt"
	"io"
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
	log.Println("Listening on " + os.Getenv("PORT"))

	for {
		conn, err := l.Accept()

		log.Println("Accepted connection")
		if err != nil {
			log.Println("Error accepting: %s", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	msg := rfc5424.Message{}
	for {
		_, err := msg.ReadFrom(conn)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("ReadFrom err: %s", err)
			return
		}

		fmt.Println(string(msg.Message))
	}
}
