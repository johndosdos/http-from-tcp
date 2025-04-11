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
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, bytesParsed)
	assert.False(t, done)

	// Test: valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("Host:   localhost:42069\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.False(t, done)

	// Test: valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.True(t, done)

	// Test: invalid header spacing
	headers = NewHeaders()
	data = []byte("    Host :  localhost:42069     \r\n\r\n")
	bytesParsed, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, bytesParsed)
	assert.False(t, done)
}
