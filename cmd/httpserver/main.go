package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/johndosdos/http-from-tcp/internal/headers"
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

func handlerRequest(w *response.Writer, req *request.Request) {
	h := headers.NewHeaders()
	h.Set("Content-Type", "text/html")
	h.Set("Connection", "close")

	switch {
	case req.RequestLine.RequestTarget == "/yourproblem":
		err := w.WriteStatusLine(response.StatusBadRequest)
		if err != nil {
			log.Printf("failed to write status line to conn: %v", err)
			return
		}

		err = w.WriteHeaders(h)

		if err != nil {
			log.Printf("failed to write headers to conn: %v", err)
			return
		}

		_, err = w.WriteBody([]byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`))
		if err != nil {
			log.Printf("failed to write body to conn: %v", err)
			return
		}

	case req.RequestLine.RequestTarget == "/myproblem":
		err := w.WriteStatusLine(response.StatusInternalServerError)
		if err != nil {
			log.Printf("failed to write status line to conn: %v", err)
			return
		}

		err = w.WriteHeaders(h)
		if err != nil {
			log.Printf("failed to write headers to conn: %v", err)
			return
		}

		_, err = w.WriteBody([]byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`))
		if err != nil {
			log.Printf("failed to write body to conn: %v", err)
			return
		}

	default:
		err := w.WriteStatusLine(response.StatusOK)
		if err != nil {
			log.Printf("failed to write status line to conn: %v", err)
			return
		}

		err = w.WriteHeaders(h)
		if err != nil {
			log.Printf("failed to write headers to conn: %v", err)
			return
		}

		_, err = w.WriteBody([]byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`))
		if err != nil {
			log.Printf("failed to write body to conn: %v", err)
			return
		}
	}
}
