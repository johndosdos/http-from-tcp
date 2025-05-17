package response

import "errors"

func (w *Writer) WriteBody(data []byte) (int, error) {
	if w.State != stateWrittenHeaders {
		return 0, errors.New("headers must be written before writing body")
	}

	w.State = stateWrittenBody
	return w.Writer.Write(data)
}
