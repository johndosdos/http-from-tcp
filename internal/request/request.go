package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	State       State
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type State int

const (
	initialized = iota
	done
)

const NUM_PARTS_REQ_LINE int = 3
const HTTP_VERSION_DIGIT string = "1.1"

func RequestFromReader(reader io.Reader) (*Request, error) {
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
