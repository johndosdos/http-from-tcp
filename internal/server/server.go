package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/johndosdos/http-from-tcp/internal/headers"
	"github.com/johndosdos/http-from-tcp/internal/request"
	"github.com/johndosdos/http-from-tcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode string
	Message    string
}

func (he *HandlerError) Write(w *response.Writer) {
	err := w.WriteStatusLine(he.StatusCode)
	if err != nil {
		log.Printf("failed to write error to conn: %v", err)
		return
	}

	var buffer bytes.Buffer

	_, err = buffer.WriteString(he.Message)
	if err != nil {
		log.Printf("failed to write message to buffer: %v", err)
		return
	}

	h := headers.NewHeaders()

	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("failed to write headers to conn: %v", err)
		return
	}

	_, err = w.Writer.Write(buffer.Bytes())
	if err != nil {
		log.Printf("failed to write response to conn: %v", err)
		return
	}
}

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("error at port %d: %w", port, err)
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}
	go server.Listen()

	return server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	return s.listener.Close()
}

func (s *Server) Listen() {
	for {
		conn, err := s.listener.Accept()
		// Check if the server is closed. There are instances where the server
		// unexpectedly shuts down after accepting a connection.
		if err != nil {
			if s.isClosed.Load() {
				log.Printf("server closed, stopped listening: %v", err)
				return
			} else {
				log.Printf("error accepting connection: %v", err)
				continue
			}
		}
		go s.Handle(conn)
	}
}

func (s *Server) Handle(conn net.Conn) {
	defer conn.Close()

	w := response.NewWriter(conn)

	parsedReq, err := request.RequestFromReader(conn)
	if err != nil {
		handlerError := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		handlerError.Write(&w)
		return
	}

	s.handler(&w, parsedReq)
}
