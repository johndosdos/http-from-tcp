package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/johndosdos/http-from-tcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusCodes := map[StatusCode]string{
		StatusOK:                  "HTTP/1.1 200 OK",
		StatusBadRequest:          "HTTP/1.1 400 Bad Request",
		StatusInternalServerError: "HTTP/1.1 500 Internal Server Error",
	}

	v, ok := statusCodes[statusCode]
	if !ok {
		return fmt.Errorf("status code not supported %v", statusCode)
	}

	statusLine := fmt.Sprintf("%v\r\n", v)
	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	return nil
}

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

	var data []byte
	for k, v := range headers {
		_, err := w.Write(fmt.Append(data, fmt.Sprintf("%v: %v\r\n", k, v)))
		if err != nil {
			return err
		}
	}

	// Write CRLF to end the headers section
	_, err := w.Write(crlf)
	if err != nil {
		return err
	}

	return nil
}
