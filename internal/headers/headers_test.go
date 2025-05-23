package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// Test: valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	bytesParsed, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, bytesParsed)
	assert.False(t, done)

	// Test: valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("Host:   localhost:42069\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.False(t, done)

	// Test: valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.True(t, done)

	// Test: lowercase header names
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.False(t, done)

	// Test: invalid header name
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NotNil(t, headers)
	require.Error(t, err)
	assert.False(t, done)

	// Test: invalid header spacing
	headers = NewHeaders()
	data = []byte("    Host :  localhost:42069     \r\n\r\n")
	bytesParsed, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, bytesParsed)
	assert.False(t, done)

	// Test: multiple header name values
	headers = NewHeaders()
	data = []byte("Accept: text/html\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NotNil(t, headers)
	require.NoError(t, err)
	assert.Equal(t, "text/html", headers["accept"])
	assert.False(t, done)

	data = []byte("Accept: application/json\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "text/html, application/json", headers["accept"])
	assert.False(t, done)
}
