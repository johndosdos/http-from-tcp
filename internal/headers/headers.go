package headers

import (
	"bytes"
	"errors"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (int, bool, error) {
	/*
		field-line = field-name ":" OWS field-value OWS

		HTTP-message = start-line CRLF
						*( field-line CRLF )
						CRLF
						[ message-body ]

		- No whitespace is allowed between the field name and colon.
		- A field line value might be preceded and/or followed by optional
			whitespace (OWS).
		- The field line value does not include that leading or trailing
			whitespace.
		- Field lines are separated by CRLF (\r\n).
		- Checking for double CRLF is a good way to detect the end of the
			field lines section (header fields).

		*From RFC 9112 section 5.1
		- A server MUST reject, with a response status code of 400 (Bad Request),
			any received request message that contains whitespace between a
			header field name and colon.
	*/

	// This function will only parse one field line (header) at a time.
	// The function will be called multiple times.

	crlf := []byte("\r\n")
	totalBytesRead := 0

	bytesRead := bytes.Index(data, crlf)
	// -1 means incomplete data
	if bytesRead == -1 {
		return 0, false, nil
	}

	/*
		Check if the first 2 bytes of the data is CRLF. This indicates the end
		of the headers section.

		bytes.Index should return 0 if CRLF is at the start of the data.

		Return true to signal successful parsing.
	*/
	if bytesRead == 0 {
		return bytesRead, true, nil
	}

	colonSep := bytes.Index(data, []byte(":"))

	fieldName := data[:colonSep]

	// Increment colonSep by 1. We want to slice the data from right after the colon
	// until bytesRead (which is at CRLF).
	fieldValue := data[colonSep+1 : bytesRead]

	// Reject if it contains whitespace between field name and colon.
	n := bytes.Index(fieldName, []byte(" "))
	if n != -1 {
		return 0, false, errors.New("bad request: invalid field name")
	}

	// Trim whitespaces at field value since they're optional.
	fieldValue = bytes.TrimSpace(fieldValue)

	h[string(fieldName)] = string(fieldValue)

	totalBytesRead = bytesRead + len(crlf)

	return totalBytesRead, false, nil
}

func NewHeaders() Headers {
	return Headers{}
}
