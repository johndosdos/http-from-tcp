package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

	case strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/stream"):
		parts := strings.Split(req.RequestLine.RequestTarget, "/")

		// We only want 4 parts right now; SP, httpbin, stream, and n. Return error if parts length is < 4.
		// SP = single space.
		if len(parts) < 4 {
			log.Printf("invalid request endpoint: %v", req.RequestLine.RequestTarget)
			return
		}

		// The last part should be n, where n is the number of requested JSON streams.
		numStreams, err := strconv.Atoi(parts[3])
		if err != nil {
			log.Printf("string to int conversion error: %v", err)
			return
		}

		url := fmt.Sprintf("https://httpbin.org/stream/%d", numStreams)

		// We serve as a proxy to the origin server.
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("failed to make request to %v: %v", url, err)
			return
		}

		defer resp.Body.Close()

		// Write status line back to client.
		err = w.WriteStatusLine(resp.Status)
		if err != nil {
			log.Printf("failed to write status line to conn: %v", err)
			return
		}

		// Write headers back to client.
		for key, values := range resp.Header {
			for _, value := range values {
				h.Set(key, value)
			}
		}

		// Manually set Transfer-Encoding header
		h.Set("Transfer-Encoding", "chunked")

		err = w.WriteHeaders(h)
		if err != nil {
			log.Printf("failed to write headers to conn: %v", err)
			return
		}

		// Forward chunked data back to client.
		data := make([]byte, 1024)

		for {
			readBytes, err := resp.Body.Read(data)
			if readBytes > 0 {
				_, err = w.WriteChunkedBody(data[:readBytes])
				if err != nil {
					log.Printf("%v", err)
					return
				}
			}

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				log.Printf("failed to read response body into p: %v", err)
				return
			}
		}

		_, err = w.WriteChunkedBodyDone()
		if err != nil {
			log.Printf("%v", err)
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
