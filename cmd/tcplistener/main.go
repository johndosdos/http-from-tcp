package main

import (
	"fmt"
	"log"
	"net"

	"github.com/johndosdos/http-from-tcp/internal/request"
)

func main() {
	addr := ":42069"
	network := "tcp"

	listener, err := net.Listen(network, addr)
	if err != nil {
		log.Fatalf("failed to listen on %s at %s: %v", network, addr, err)
	}
	defer listener.Close()

	fmt.Printf("Server started. Listening on \"%s\" at \"%s\"\n", network, addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection from %s", conn.LocalAddr())
		}

		fmt.Printf("CONNECTION ACCEPTED...\n\n")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error processing connection: %v", err)
		}

		fmt.Printf(`Request line:
- Method: %v
- Target: %v
- Version: %v
`, req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %v: %v\n", key, value)
		}

		fmt.Println("Body:")
		fmt.Printf("%s", string(req.Body))

		fmt.Printf("\n...CONNECTION CLOSED\n")
	}
	/* 	file, err := os.Open("messages.txt")
	   	handleError(err, "error opening file")
	   	defer file.Close() */

}
