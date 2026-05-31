package atomic

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressReader_CountsBytes(t *testing.T) {
	data := []byte("hello world — bifrost progress test")
	var counted int64
	r := &progressReader{r: bytes.NewReader(data), fn: func(n int64) { counted += n }}

	_, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, int64(len(data)), counted)
}

func TestProgressReader_CountsAcrossMultipleReads(t *testing.T) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}

	var counted int64
	r := &progressReader{r: bytes.NewReader(data), fn: func(n int64) { counted += n }}

	buf := make([]byte, 100)
	for {
		_, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}
	assert.Equal(t, int64(len(data)), counted)
}
