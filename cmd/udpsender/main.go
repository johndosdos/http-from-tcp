package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("failed resolve server address: %v\n", err)
	}

	udpConn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatalf("failed to establish UDP connection: %v\n", err)
	}
	defer udpConn.Close()

	bufReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := bufReader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("end of input reached: %v\n", err)
			} else {
				log.Printf("error reading from input: %v\n", err)
			}
		}

		_, err = udpConn.Write([]byte(input))
		if err != nil {
			log.Printf("error writing to UDP: %v", err)
		}
	}
}
