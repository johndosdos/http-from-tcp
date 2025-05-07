package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/johndosdos/http-from-tcp/internal/response"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("error at port %d: %w", port, err)
	}

	server := &Server{
		listener: listener,
	}
	go server.Listen()

	return server, nil
}

func (s *Server) Close() error {
	if !s.isClosed.Load() {
		err := s.listener.Close()
		if err != nil {
			return fmt.Errorf("there's a problem closing the server: %w", err)
		}
	}
	s.isClosed.Store(true)

	return nil
}

func (s *Server) Listen() {
	for {
		conn, err := s.listener.Accept()
		// Check if the server is closed. There are instances where the server
		// unexpectedly shuts down after accepting a connection.
		if err != nil {
			if s.isClosed.Load() {
				log.Printf("server closed, stopped listening: %v", err)
				break
			} else {
				log.Printf("error accepting connection: %v", err)
				break
			}
		}
		go s.Handle(conn)
	}
}

func (s *Server) Handle(conn net.Conn) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)
	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Printf("error writing status line: %v", err)
	}

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Printf("error writing headers field: %v", err)
	}
}
