package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/johndosdos/http-from-tcp/internal/request"
	"github.com/johndosdos/http-from-tcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (he *HandlerError) Write(w io.Writer) {
	err := response.WriteStatusLine(w, he.StatusCode)
	if err != nil {
		fmt.Printf("failed to write status-line: %v", err)
		return
	}

	var buffer bytes.Buffer

	_, err = buffer.WriteString(he.Message)
	if err != nil {
		fmt.Printf("failed to write message to buffer: %v", err)
		return
	}

	h := response.GetDefaultHeaders(buffer.Len())

	err = response.WriteHeaders(w, h)
	if err != nil {
		fmt.Printf("failed to write headers: %v", err)
		return
	}

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		fmt.Printf("failed to write response to conn: %v", err)
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

	parsedReq, err := request.RequestFromReader(conn)
	if err != nil {

		handlerError := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		handlerError.Write(conn)
		return
	}

	var buffer bytes.Buffer

	handlerErr := s.handler(&buffer, parsedReq)
	if handlerErr != nil {
		handlerErr.Write(conn)
		return
	}

	contentLength := buffer.Len()
	headers := response.GetDefaultHeaders(contentLength)

	err = response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Printf("error writing status line: %v", err)
		return
	}

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Printf("error writing headers field: %v", err)
		return
	}

	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		log.Println(err)
		return
	}
}
