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

func (c *Canvas) String() string {
	var buffer bytes.Buffer

	for h := 0; h < c.Height; h++ {
		for w := 0; w < c.Width; w++ {
			idx := Index{w, h}

			// Search 'text' map; if nothing there try the 'data' map.
			r, ok := c.text[idx]
			if !ok {
				r = c.runeAt(idx)
			}

			_, err := buffer.WriteRune(r)
			if err != nil {
				continue
			}
		}

		err := buffer.WriteByte('\n')
		if err != nil {
			continue
		}
	}

	return buffer.String()
}
