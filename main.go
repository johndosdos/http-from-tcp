package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	file, err := os.Open("messages.txt")
	handleError(err, "error opening file")
	defer file.Close()

	buf := make([]byte, 8)

	for {
		_, err = file.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			handleError(err, "error reading from file into buffer")
		}

		fmt.Printf("read: %s\n", buf)
	}
}
