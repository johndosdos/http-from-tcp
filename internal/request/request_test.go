package request

import (
	"io"
	"testing"

	"github.com/johndosdos/http-from-tcp/internal/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}

func TestRequest(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestRequestFromReader_Headers(t *testing.T) {
	// Test: standard headers
	reader := &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	req, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, "localhost:42069", req.Headers["host"])
	assert.Equal(t, "curl/7.81.0", req.Headers["user-agent"])
	assert.Equal(t, "*/*", req.Headers["accept"])

	// Test: empty headers
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}
	req, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, headers.Headers{}, req.Headers)

	// Test: malformed headers
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	req, err = RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, req)

	// Test: duplicate headers
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nHost: localhost:42069\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	req, err = RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, req)
	// assert.Equal(t, headers.Headers{}, req.Headers)

	// Test: case-insensitive headers
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHOST: localhost:42069\r\nUSER-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	req, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, "localhost:42069", req.Headers["host"])
	assert.Equal(t, "curl/7.81.0", req.Headers["user-agent"])
	assert.Equal(t, "*/*", req.Headers["accept"])

	// Test: missing end of headers
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n",
		numBytesPerRead: 3,
	}
	req, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, "localhost:42069", req.Headers["host"])
	assert.Equal(t, "curl/7.81.0", req.Headers["user-agent"])
	assert.Equal(t, "*/*", req.Headers["accept"])

	// Test: invalid header characters
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nH()st: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	req, err = RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, req)
}
