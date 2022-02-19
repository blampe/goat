package goat

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"
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
	c := qt.New(t)

	var buf safeBuffer

	// TODO: UNICODE
	buf.WriteString(" +-->\n")
	buf.WriteString(" | å\n")
	buf.WriteString(" +----->")

	canvas := NewCanvas(bytes.NewReader(buf.Bytes()))

	c.Assert(canvas.Width, qt.Equals, 8)
	c.Assert(canvas.Height, qt.Equals, 3)

	buf.Truncate(0)
	buf.WriteString(" +-->   \n")
	buf.WriteString(" | å    \n")
	buf.WriteString(" +----->\n")

	expected := buf.String()

	c.Assert(expected, qt.Equals, canvas.String())
}
