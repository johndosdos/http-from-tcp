package response

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/johndosdos/http-from-tcp/internal/headers"
	"github.com/johndosdos/http-from-tcp/internal/request"
)

func HandlerProxy(w *Writer, req *request.Request, h headers.Headers) error {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")

	url := fmt.Sprintf("https://httpbin.org/%s", target)

	// We serve as a proxy to the origin server.
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make request to %v: %v", url, err)
	}

	defer resp.Body.Close()

	// Write status line back to client.
	err = w.WriteStatusLine(resp.Status)
	if err != nil {
		return fmt.Errorf("failed to write status line to conn: %v", err)
	}

	// Write headers back to client.
	for key, values := range resp.Header {
		for _, value := range values {
			h.Set(key, value)
		}
	}

	// Manually set Transfer-Encoding header
	h.Set("Transfer-Encoding", "chunked")

	// Set Trailers.
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")

	err = w.WriteHeaders(h)
	if err != nil {
		return fmt.Errorf("failed to write headers to conn: %v", err)
	}

	// Forward chunked data back to client.
	data := make([]byte, 1024)

	// We need to keep track of the full response body so that we can
	// calculate its hash and its length.
	hasher := sha256.New()
	totalDataLen := 0

	for {
		readBytes, err := resp.Body.Read(data)
		if readBytes > 0 {
			_, err = w.WriteChunkedBody(data[:readBytes])
			if err != nil {
				return fmt.Errorf("failed to write chunked data: %v", err)
			}

			hasher.Write(data[:readBytes])
			totalDataLen += readBytes
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to read response body into p: %v", err)
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		return fmt.Errorf("failed to write last chunked data: %v", err)
	}

	// Calculate hash and length of the response data.
	trailerHash := hasher.Sum(nil)
	trailerLen := totalDataLen

	h.Set("X-Content-SHA256", fmt.Sprintf("%x", trailerHash))
	h.Set("X-Content-Length", fmt.Sprintf("%d", trailerLen))

	err = w.WriteTrailers(h)
	if err != nil {
		return fmt.Errorf("failed to write Trailers: %v", err)
	}

	return nil
}
