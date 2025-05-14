package response

import (
	"fmt"
	"io"
)

type StatusCode string

const (
	StatusOK                  StatusCode = "200"
	StatusBadRequest          StatusCode = "400"
	StatusInternalServerError StatusCode = "500"
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reasonPhrase := ""

	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %v %v\r\n", statusCode, reasonPhrase)
	_, err := w.Write([]byte(statusLine))
	return err
}
