package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
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

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hello World!\r\n"
	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("error writing data to connection: %v", err)
	}
}
