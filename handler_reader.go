package main

import (
	"errors"
	"io"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	msgChannel := make(chan string)

	go func() {
		defer close(msgChannel)

		buffer := make([]byte, 8)
		var stringBuffer strings.Builder

		for {
			bytesRead, err := f.Read(buffer)
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
			msgChannel <- line
		}
	}()

	return msgChannel
}
