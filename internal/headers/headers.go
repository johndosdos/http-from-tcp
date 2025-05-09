package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

func (h Headers) Get(key string) string {
	return h[key]
}

func (h Headers) Set(key, value string) {
	// Check if header name already exists in the map.
	// If header name exists but both have different values, join them.
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{v, value}, ", ")
	}

	h[key] = value
}

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
		totalBytesRead += len(crlf)
		return totalBytesRead, true, nil
	}

	colonSep := bytes.Index(data, []byte(":"))
	if colonSep == -1 {
		return 0, false, errors.New("bad request: malformed field name")
	}

	fieldName := data[:colonSep]

	// Check field name for invalid chars. Return an error if so.
	// Reject if it contains whitespace between field name and colon.
	if ok := isHeaderNameValid(fieldName); !ok {
		return 0, false, errors.New("bad request: invalid field name")
	}

	// Increment colonSep by 1. We want to slice the data from right after the colon
	// until bytesRead (which is at CRLF).
	fieldValue := data[colonSep+1 : bytesRead]

	// Trim whitespaces at field value since they're optional.
	fieldValue = bytes.TrimSpace(fieldValue)

	// Covert header name and value chars to lowercase.
	fieldName = bytes.ToLower(fieldName)
	fieldValue = bytes.ToLower(fieldValue)

	fieldNameStr := string(fieldName)
	fieldValueStr := string(fieldValue)

	// If header name exists but have the same value, reject it.
	if _, ok := h[fieldNameStr]; ok && fieldNameStr == "host" {
		return totalBytesRead, false, errors.New("bad request: only one host header allowed")
	}

	h.Set(fieldNameStr, fieldValueStr)

	totalBytesRead = bytesRead + len(crlf)

	return totalBytesRead, false, nil
}

func NewHeaders() Headers {
	return Headers{}
}

func isHeaderNameValid(headerName []byte) bool {
	const allowed = "!#$%&'*+-.^_`|~"
	var ok bool

	for _, char := range headerName {
		if ('A' <= char && char <= 'Z') ||
			('a' <= char && char <= 'z') ||
			('0' <= char && char <= '9') {
			ok = true
		} else if bytes.ContainsRune([]byte(allowed), rune(char)) {
			ok = true
		} else {
			ok = false
			break
		}
	}

	return ok
}
