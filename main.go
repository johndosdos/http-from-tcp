package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	handleError(err, "error opening file")
	defer file.Close()

	buffer := make([]byte, 8)
	var stringBuffer strings.Builder

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			handleError(err, "error reading from file into buffer")
		}

		// file.Read() returns (0, io.EOF) at EOF.
		if bytesRead > 0 {
			_, err := stringBuffer.Write(buffer[:bytesRead])
			handleError(err, "error writing string into buffer")
		}
	}

	lines := strings.SplitSeq(stringBuffer.String(), "\n")

	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}
