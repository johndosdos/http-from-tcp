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
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	requestData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to read contents to memory: %w", err)
	}

	requestLine, err := parseRequestLine(requestData)
	if err != nil {
		return nil, fmt.Errorf("unable to parse request-line: %w", err)
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

func parseRequestLine(requestData []byte) (*RequestLine, error) {
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

	const NUM_PARTS_REQ_LINE int = 3
	const HTTP_VERSION string = "1.1"

	// There are four parts to an HTTP Message. But since we're splitting
	// at CRLF, then there are three parts.

	// Trim excess whitespace.
	httpMsgParts := bytes.Split(bytes.TrimSpace(requestData), []byte("\r\n"))

	// Verify the parts of the request-line.
	requestLine := bytes.Split(httpMsgParts[0], []byte(" "))
	if len(requestLine) != NUM_PARTS_REQ_LINE {
		return nil, fmt.Errorf("invalid HTTP message request-line: %d", len(requestLine))
	}

	reqMethod := string(requestLine[0])
	reqTarget := string(requestLine[1])

	// extract the digit part from HTTP-version
	reqHTTPVersion := bytes.Split(requestLine[2], []byte("/"))
	httpVersionDigit := reqHTTPVersion[1]

	// Verify request-line method to have uppercase chars.
	for _, char := range reqMethod {
		if !unicode.IsUpper(char) {
			return nil, fmt.Errorf("invalid HTTP method: received: '%s', expected: '%s'", reqMethod, strings.ToUpper(reqMethod))
		}
	}

	// Verify HTTP-version. We only allow HTTP/1.1.
	if string(httpVersionDigit) != HTTP_VERSION {
		return nil, fmt.Errorf("invalid request-line HTTP-version; should be %s: %v", HTTP_VERSION, reqHTTPVersion)
	}

	return &RequestLine{
		Method:        reqMethod,
		RequestTarget: reqTarget,
		HttpVersion:   string(httpVersionDigit),
	}, nil
}
