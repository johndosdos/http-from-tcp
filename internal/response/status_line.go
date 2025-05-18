package response

import (
	"errors"
	"fmt"
)

type StatusCode string

const (
	StatusOK                  StatusCode = "200"
	StatusBadRequest          StatusCode = "400"
	StatusInternalServerError StatusCode = "500"
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != stateInit {
		return errors.New("status line has already been written")
	}

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
	_, err := w.Writer.Write([]byte(statusLine))

	w.State = stateWrittenStatusLine

	return err
}
