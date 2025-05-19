package response

import (
	"errors"
	"fmt"
)

const (
	StatusOK                  string = "200"
	StatusBadRequest          string = "400"
	StatusInternalServerError string = "500"
)

func (w *Writer) WriteStatusLine(statusCode string) error {
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
