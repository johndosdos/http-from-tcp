package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/johndosdos/http-from-tcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       State
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type State int

const (
	INITIALIZED = iota
	DONE
	REQUEST_STATE_PARSING_HEADERS
)

const NUM_PARTS_REQ_LINE int = 3
const HTTP_VERSION_DIGIT string = "1.1"

const BUFFER_SIZE int = 8

func (r *Request) Parse(data []byte) (int, error) {
	switch r.State {
	case INITIALIZED:
		requestLine, bytesRead, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if bytesRead == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.State = REQUEST_STATE_PARSING_HEADERS
		return bytesRead, nil
	case REQUEST_STATE_PARSING_HEADERS:
		totalBytesParsed := 0
		isHeaderDone := false

		if r.Headers == nil {
			r.Headers = headers.NewHeaders()
		}

		for !isHeaderDone {
			bytesParsed, done, err := r.Headers.Parse(data[totalBytesParsed:])
			if err != nil {
				return 0, err
			}

			if bytesParsed == 0 && !done {
				return totalBytesParsed, err
			}

			isHeaderDone = done
			totalBytesParsed += bytesParsed
		}

		r.State = DONE
		return totalBytesParsed, nil
	case DONE:
		return 0, errors.New("error: trying to read data in 'done' state")
	}

	return 0, fmt.Errorf("error: parser encountered unknown state: %v", r.State)
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	// Track how many bytes have we read from the io.Reader (request data) into
	// the buffer.
	bytesInBuffer := 0

	request := &Request{
		State: INITIALIZED,
	}

	buffer := make([]byte, BUFFER_SIZE)

	for request.State != DONE {
		if bytesInBuffer == cap(buffer) {
			newBuffer := make([]byte, cap(buffer)*2)
			copy(newBuffer, buffer[:bytesInBuffer])
			buffer = newBuffer
		}

		bytesRead, err := reader.Read(buffer[bytesInBuffer:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
		}

		bytesInBuffer += bytesRead

		bytesParsed, err := request.Parse(buffer[:bytesInBuffer])
		if err != nil {
			return nil, fmt.Errorf("unable to parse request data: %w", err)
		}

		copy(buffer[0:], buffer[bytesParsed:bytesInBuffer])
		bytesInBuffer -= bytesParsed
	}

	return request, nil
	/*
		 	requestData, err := io.ReadAll(reader)
			if err != nil {
				return nil, fmt.Errorf("unable to read contents to memory: %w", err)
			}

			requestLine, bytesRead, err := parseRequestLine(requestData)
			if err != nil {
				return nil, fmt.Errorf("unable to parse request-line: %w", err)
			}

			return &Request{
				RequestLine: *requestLine,
				State:       initialized,
			}, nil
	*/
}

func parseRequestLine(requestData []byte) (*RequestLine, int, error) {
	/*
		HTTP-message = start-line CRLF			<--- either request-line or status-line
						*( field-line CRLF )	<--- header/s
						CRLF					<--- carriage return line feed
						[ message-body ]		<--- optional body
	*/

	/*
		request-line   = method SP request-target SP HTTP-version
		where:
				method 			<--- HTTP Methods (e.g., POST, GET, PATCH, PUT, DELETE, etc.)
				SP 				<--- Single Space
				request-target 	<--- /path from GET /path/to/resource?=query HTTP/1.1
				HTTP-version	<--- HTTP-name "/" DIGIT "." DIGIT (e.g., HTTP/1.1)
	*/

	// CRLF is 2 bytes, \r and \n.
	crlf := []byte("\r\n")
	totalBytesRead := 0

	// If crlf is not found in the request data, i.e., incomplete data.
	bytesRead := bytes.Index(requestData, crlf)
	if bytesRead == -1 {
		return nil, 0, nil
	}

	totalBytesRead += bytesRead + len(crlf)

	// Extract request-line from the HTTP message.
	requestLine := requestData[:bytesRead]
	parts := bytes.Split(requestLine, []byte(" "))
	reqMethod := parts[0]
	reqTarget := parts[1]
	reqHTTPVersion := parts[2]

	// extract the digit part from HTTP-version
	httpVerParts := bytes.Split(reqHTTPVersion, []byte("/"))
	httpVerDigit := string(httpVerParts[1])

	// Verify request-line method to have uppercase chars.
	if !verifyMethod(reqMethod) {
		return nil, bytesRead, fmt.Errorf("invalid HTTP method: received: '%s', expected: '%s'", reqMethod, strings.ToUpper(string(reqMethod)))
	}

	// Verify HTTP-version. We only allow HTTP/1.1.
	if !verifyVersion(HTTP_VERSION_DIGIT, httpVerDigit) {
		return nil, bytesRead, fmt.Errorf("invalid HTTP version; received: '%s', expected: '%s'", httpVerDigit, HTTP_VERSION_DIGIT)
	}

	return &RequestLine{
		Method:        string(reqMethod),
		RequestTarget: string(reqTarget),
		HttpVersion:   httpVerDigit,
	}, totalBytesRead, nil
}

func verifyMethod(method []byte) bool {
	for _, char := range method {
		if !unicode.IsUpper(rune(char)) {
			return false
		}
	}

	return true
}

func verifyVersion(ref, actual string) bool {
	return actual == ref
}
