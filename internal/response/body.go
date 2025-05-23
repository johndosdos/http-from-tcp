package response

import (
	"errors"
	"fmt"
	"io"

	"github.com/johndosdos/http-from-tcp/internal/headers"
)

func (w *Writer) WriteBody(data []byte) (int, error) {
	if w.State != stateWrittenHeaders {
		return 0, errors.New("headers must be written before writing body")
	}

	w.State = stateWrittenBody
	return w.Writer.Write(data)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	// Refer to RFC 9112 7.1
	totalBytesWritten := 0

	// Write chunk size to conn first.
	chunkSize := fmt.Sprintf("%x\r\n", len(p))

	n, err := io.WriteString(w.Writer, chunkSize)
	if err != nil {
		return 0, err
	}

	totalBytesWritten += n

	// Followed by the chunk data
	chunkData := append(p, '\r', '\n')
	n, err = w.Writer.Write(chunkData)
	if err != nil {
		return 0, err
	}

	return totalBytesWritten + n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	// Write the final chunked data section
	totalBytesWritten := 0

	lastChunk := fmt.Sprintf("%d\r\n", 0)
	n, err := io.WriteString(w.Writer, lastChunk)
	if err != nil {
		return 0, err
	}

	totalBytesWritten += n

	return totalBytesWritten + n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	crlf := []byte("\r\n")

	for k, v := range h {
		headerLine := fmt.Sprintf("%v: %v\r\n", k, v)
		_, err := w.Writer.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}

	// Write CRLF to end the trailers section
	_, err := w.Writer.Write(crlf)
	return err
}
