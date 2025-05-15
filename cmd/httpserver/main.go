package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/johndosdos/http-from-tcp/internal/request"
	"github.com/johndosdos/http-from-tcp/internal/response"
	"github.com/johndosdos/http-from-tcp/internal/server"
)

func main() {
	const port = 42069

	server, err := server.Serve(port, handlerRequest)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handlerRequest(w io.Writer, req *request.Request) *server.HandlerError {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		return &server.HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    "Your problem is not my problem\n",
		}
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		return &server.HandlerError{
			StatusCode: response.StatusInternalServerError,
			Message:    "Woopsie, my bad\n",
		}
	}

	_, err := io.WriteString(w, "All good, frfr\n")
	if err != nil {
		return &server.HandlerError{
			StatusCode: response.StatusInternalServerError,
			Message:    "an error occurred while writing the response.\n",
		}
	}

	return nil
}
