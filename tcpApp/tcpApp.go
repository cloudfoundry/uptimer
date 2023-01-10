package tcpApp

const Source = `package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("Listening on 0.0.0.0:8080")

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
    defer conn.Close()
    remoteAddr := conn.RemoteAddr()
    fmt.Printf("Remote Address: %s\n", remoteAddr)

    // Make a buffer to hold incoming data.
    buff := make([]byte, 1024)
    // Continue to receive the data forever...
    for {
        // Read the incoming connection into the buffer.
        readBytes, err := conn.Read(buff)
        if err != nil {
            fmt.Printf("Closing connection to %s: %s\n", remoteAddr, err.Error())
            return
        }
        var writeBuffer bytes.Buffer
        writeBuffer.WriteString("uptimer")
        writeBuffer.WriteString(":")
        writeBuffer.Write(buff[0:readBytes])
        fmt.Printf("Message to %s: %s\n", remoteAddr, writeBuffer.String())
        conn.Write([]byte("Hello from Uptimer.\n"))
        if err != nil {
            fmt.Printf("Closing connection to %s: %s\n", remoteAddr, err.Error())
            return
        }
    }
}
`
