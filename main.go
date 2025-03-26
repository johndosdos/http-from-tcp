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

	buf := make([]byte, 8)
	var strBuf strings.Builder

	for {
		_, err = file.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			handleError(err, "error reading from file into buffer")
		}

		// fmt.Printf("read: %s\n", buf)
		_, err := strBuf.Write(buf)
		handleError(err, "error writing to string buffer")
	}

	// processedStr := strings.Split(strBuf.String(), "\n")
	lines := strings.SplitSeq(strBuf.String(), "\n")
	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}
