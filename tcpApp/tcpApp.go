package tcpApp

import (
	"fmt"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", os.Getenv("TCP_PORT")))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("Listening on localhost:%s\n", os.Getenv("TCP_PORT"))

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	conn.Write([]byte("Message received.\n"))
	conn.Close()
}
