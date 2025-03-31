package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("messages.txt")
	handleError(err, "error opening file")
	defer file.Close()

	lineChannel := getLinesChannel(file)

	for line := range lineChannel {
		fmt.Printf("read: %s\n", line)
	}
}
