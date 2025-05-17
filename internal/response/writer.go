package response

import (
	"io"
	"net"
)

type writerState int

type Writer struct {
	Writer io.Writer
	State  writerState
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
