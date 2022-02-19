package goat

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// SafeBuffer is intended only for testing and doesn't check errors on writes
// (to make the errcheck linter happy).
type safeBuffer struct {
	bytes.Buffer
}

func (b *safeBuffer) WriteString(s string) {
	_, err := b.Buffer.WriteString(s)

	if err != nil {
		panic(err)
	}
}

func TestReadASCII(t *testing.T) {

	var buf safeBuffer

	// TODO: UNICODE
	buf.WriteString(" +-->\n")
	buf.WriteString(" | å\n")
	buf.WriteString(" +----->")

	canvas := NewCanvas(bytes.NewReader(buf.Bytes()))

	assert.Equal(t, 8, canvas.Width)
	assert.Equal(t, 3, canvas.Height)

	buf.Truncate(0)
	buf.WriteString(" +-->   \n")
	buf.WriteString(" | å    \n")
	buf.WriteString(" +----->\n")

	expected := buf.String()

	assert.Equal(t, expected, canvas.String())
}
