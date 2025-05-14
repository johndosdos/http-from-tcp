package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/johndosdos/http-from-tcp/internal/headers"
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	contentLenStr := strconv.Itoa(contentLen)

	h.Set("Content-Length", contentLenStr)
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	crlf := []byte("\r\n")

	for k, v := range headers {
		headerLine := fmt.Sprintf("%v: %v\r\n", k, v)
		_, err := w.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}

	// Write CRLF to end the headers section
	_, err := w.Write(crlf)
	return err
}
