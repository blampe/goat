package goat

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestReadASCII(t *testing.T) {
	c := qt.New(t)

	var buf bytes.Buffer

	// TODO: UNICODE
	buf.WriteString(" +-->\n")
	buf.WriteString(" | å\n")
	buf.WriteString(" +----->")

	canvas := NewCanvas(&buf)

	c.Assert(canvas.Width, qt.Equals, 8)
	c.Assert(canvas.Height, qt.Equals, 3)

	buf.Reset()
	buf.WriteString(" +-->   \n")
	buf.WriteString(" | å    \n")
	buf.WriteString(" +----->\n")

	expected := buf.String()

	c.Assert(expected, qt.Equals, canvas.String())
}
