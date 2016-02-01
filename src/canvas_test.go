package goaat

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadASCII(t *testing.T) {

	var buf bytes.Buffer

	// TODO: UNICODE
	buf.WriteString(" +-->\n")
	buf.WriteString(" |\n")
	buf.WriteString(" +----->")

	canvas := NewCanvas(bytes.NewReader(buf.Bytes()))

	assert.Equal(t, 8, canvas.Width)
	assert.Equal(t, 3, canvas.Height)

	buf.Truncate(0)
	buf.WriteString(" +-->   \n")
	buf.WriteString(" |      \n")
	buf.WriteString(" +----->\n")

	expected := buf.String()

	assert.Equal(t, expected, canvas.String())
}
