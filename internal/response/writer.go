package response

import (
	"io"
	"net"
)

type Writer struct {
	Writer io.Writer
	State  int
}

const (
	stateInit = iota
	stateWrittenStatusLine
	stateWrittenHeaders
	stateWrittenBody
)

func NewWriter(w net.Conn) Writer {
	return Writer{
		Writer: w,
		State:  stateInit,
	}
}
