package response

import (
	"errors"
	"fmt"

	"github.com/johndosdos/http-from-tcp/internal/headers"
)

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != stateWrittenStatusLine {
		return errors.New("status line must be written before writing headers")
	}

	crlf := []byte("\r\n")

	for k, v := range headers {
		headerLine := fmt.Sprintf("%v: %v\r\n", k, v)
		_, err := w.Writer.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}

	// Write CRLF to end the headers section
	_, err := w.Writer.Write(crlf)
	w.State = stateWrittenHeaders
	return err
}
